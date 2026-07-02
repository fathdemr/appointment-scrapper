package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"appointment-scrapper/config"
	"appointment-scrapper/internal/browser"
	"appointment-scrapper/internal/notifier"
	"appointment-scrapper/internal/scheduler"
	"appointment-scrapper/internal/scraper"
	"appointment-scrapper/model"
	"appointment-scrapper/repository"
)

// CreateJobRequest yeni bir booking job oluşturmak için istek verisi.
type CreateJobRequest struct {
	Name          string   `json:"name" binding:"required"`
	TCNo          string   `json:"tc_no" binding:"required"`
	Password      string   `json:"password" binding:"required"`
	SportType     string   `json:"sport_type" binding:"required"`
	Facilities    []string `json:"facilities" binding:"required"`
	Courts        []string `json:"courts"`
	TargetDates   []string `json:"target_dates" binding:"required"`
	DesiredTimes  []string `json:"desired_times" binding:"required"`

	TelegramBotToken     string `json:"telegram_bot_token"`
	TelegramChatID       string `json:"telegram_chat_id"`
	SuccessMessage       string `json:"success_message"`

	OpeningOffsetHours    int  `json:"opening_offset_hours"`
	BurstBeforeSeconds    int  `json:"burst_before_seconds"`
	BurstAfterSeconds     int  `json:"burst_after_seconds"`
	BurstIntervalSeconds  int  `json:"burst_interval_seconds"`
	PollIntervalSeconds   int  `json:"poll_interval_seconds"`
	Headless              bool `json:"headless"`
	BrowserTimeoutSeconds int  `json:"browser_timeout_seconds"`
}

// UpdateJobRequest mevcut bir job'ı güncellemek için istek verisi.
// Şifre boş bırakılırsa mevcut şifre korunur.
type UpdateJobRequest struct {
	Name          string   `json:"name" binding:"required"`
	TCNo          string   `json:"tc_no" binding:"required"`
	Password      string   `json:"password"`
	SportType     string   `json:"sport_type" binding:"required"`
	Facilities    []string `json:"facilities" binding:"required"`
	Courts        []string `json:"courts"`
	TargetDates   []string `json:"target_dates" binding:"required"`
	DesiredTimes  []string `json:"desired_times" binding:"required"`

	TelegramBotToken     string `json:"telegram_bot_token"`
	TelegramChatID       string `json:"telegram_chat_id"`
	SuccessMessage       string `json:"success_message"`

	OpeningOffsetHours    int  `json:"opening_offset_hours"`
	BurstBeforeSeconds    int  `json:"burst_before_seconds"`
	BurstAfterSeconds     int  `json:"burst_after_seconds"`
	BurstIntervalSeconds  int  `json:"burst_interval_seconds"`
	PollIntervalSeconds   int  `json:"poll_interval_seconds"`
	Headless              bool `json:"headless"`
	BrowserTimeoutSeconds int  `json:"browser_timeout_seconds"`
}

// JobStatusResponse DB durumu ile anlık zamanlayıcı durumunu birleştirir.
type JobStatusResponse struct {
	*model.BookingJob
	SchedulerMode string     `json:"scheduler_mode,omitempty"`
	NextBurstAt   *time.Time `json:"next_burst_at,omitempty"`
	NextBurstSlot string     `json:"next_burst_slot,omitempty"`
	SMSWaiting    bool       `json:"sms_waiting"`
	IsRunning     bool       `json:"is_running"`
}

// JobService booking job'larının yaşam döngüsünü yönetir.
type JobService interface {
	CreateJob(ctx context.Context, req CreateJobRequest) (*model.BookingJob, error)
	GetJob(ctx context.Context, id string) (*model.BookingJob, error)
	ListJobs(ctx context.Context) ([]*model.BookingJob, error)
	UpdateJob(ctx context.Context, id string, req UpdateJobRequest) (*model.BookingJob, error)
	DeleteJob(ctx context.Context, id string) error

	StartJob(ctx context.Context, id string) error
	StopJob(ctx context.Context, id string) error
	RunJobNow(ctx context.Context, id string) error

	VerifyTelegram(ctx context.Context, id string) error
	SubmitSMSCode(jobID string, code string) bool

	GetJobStatus(ctx context.Context, id string) (*JobStatusResponse, error)
	RestoreActiveJobs(ctx context.Context) error
}

type jobEntry struct {
	sched *scheduler.Scheduler
}

type jobServiceImpl struct {
	repo     repository.JobRepository
	registry *SMSCodeRegistry
	logger   *zap.Logger

	mu   sync.RWMutex
	jobs map[string]*jobEntry // jobID → çalışan zamanlayıcı
}

func NewJobService(repo repository.JobRepository, registry *SMSCodeRegistry, logger *zap.Logger) JobService {
	return &jobServiceImpl{
		repo:     repo,
		registry: registry,
		logger:   logger,
		jobs:     make(map[string]*jobEntry),
	}
}

// ─── CRUD ────────────────────────────────────────────────────────────────────

func (s *jobServiceImpl) CreateJob(ctx context.Context, req CreateJobRequest) (*model.BookingJob, error) {
	job := &model.BookingJob{
		Name:                 req.Name,
		Status:               model.JobStatusPending,
		TCNo:                 req.TCNo,
		Password:             req.Password,
		SportType:            req.SportType,
		Facilities:           req.Facilities,
		Courts:               req.Courts,
		TargetDates:          req.TargetDates,
		DesiredTimes:         req.DesiredTimes,
		TelegramBotToken:     req.TelegramBotToken,
		TelegramChatID:       req.TelegramChatID,
		SuccessMessage:       req.SuccessMessage,
		OpeningOffsetHours:   req.OpeningOffsetHours,
		BurstBeforeSeconds:   req.BurstBeforeSeconds,
		BurstAfterSeconds:    req.BurstAfterSeconds,
		BurstIntervalSeconds: req.BurstIntervalSeconds,
		PollIntervalSeconds:  req.PollIntervalSeconds,
		Headless:             req.Headless,
		BrowserTimeoutSeconds: req.BrowserTimeoutSeconds,
	}
	applyDefaults(job)
	if err := s.repo.Create(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *jobServiceImpl) GetJob(ctx context.Context, id string) (*model.BookingJob, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *jobServiceImpl) ListJobs(ctx context.Context) ([]*model.BookingJob, error) {
	return s.repo.List(ctx)
}

func (s *jobServiceImpl) UpdateJob(ctx context.Context, id string, req UpdateJobRequest) (*model.BookingJob, error) {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	_, running := s.jobs[id]
	s.mu.RUnlock()
	if running {
		return nil, fmt.Errorf("çalışan bir job güncellenemez; önce durdurun")
	}

	job.Name = req.Name
	job.TCNo = req.TCNo
	if req.Password != "" {
		job.Password = req.Password
	}
	job.SportType = req.SportType
	job.Facilities = req.Facilities
	job.Courts = req.Courts
	job.TargetDates = req.TargetDates
	job.DesiredTimes = req.DesiredTimes
	job.TelegramBotToken = req.TelegramBotToken
	job.TelegramChatID = req.TelegramChatID
	job.SuccessMessage = req.SuccessMessage
	job.OpeningOffsetHours = req.OpeningOffsetHours
	job.BurstBeforeSeconds = req.BurstBeforeSeconds
	job.BurstAfterSeconds = req.BurstAfterSeconds
	job.BurstIntervalSeconds = req.BurstIntervalSeconds
	job.PollIntervalSeconds = req.PollIntervalSeconds
	job.Headless = req.Headless
	job.BrowserTimeoutSeconds = req.BrowserTimeoutSeconds
	applyDefaults(job)

	if err := s.repo.Update(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *jobServiceImpl) DeleteJob(ctx context.Context, id string) error {
	s.mu.Lock()
	if entry, ok := s.jobs[id]; ok {
		entry.sched.Stop()
		delete(s.jobs, id)
	}
	s.mu.Unlock()
	return s.repo.Delete(ctx, id)
}

// ─── Yaşam Döngüsü ───────────────────────────────────────────────────────────

func (s *jobServiceImpl) StartJob(ctx context.Context, id string) error {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	s.mu.RLock()
	_, running := s.jobs[id]
	s.mu.RUnlock()
	if running {
		return fmt.Errorf("job zaten çalışıyor")
	}

	s.startJobInternal(job)
	return s.repo.UpdateStatus(ctx, id, model.JobStatusActive)
}

func (s *jobServiceImpl) StopJob(ctx context.Context, id string) error {
	s.mu.Lock()
	entry, ok := s.jobs[id]
	if ok {
		entry.sched.Stop()
		delete(s.jobs, id)
	}
	s.mu.Unlock()
	return s.repo.UpdateStatus(ctx, id, model.JobStatusStopped)
}

func (s *jobServiceImpl) RunJobNow(ctx context.Context, id string) error {
	s.mu.RLock()
	entry, ok := s.jobs[id]
	s.mu.RUnlock()

	if !ok {
		job, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		s.startJobInternal(job)
		_ = s.repo.UpdateStatus(ctx, id, model.JobStatusActive)
		s.mu.RLock()
		entry = s.jobs[id]
		s.mu.RUnlock()
	}

	entry.sched.RunNow()
	return nil
}

// ─── Telegram & SMS ──────────────────────────────────────────────────────────

func (s *jobServiceImpl) VerifyTelegram(ctx context.Context, id string) error {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if job.TelegramBotToken == "" || job.TelegramChatID == "" {
		return errors.New("telegram yapılandırılmamış")
	}
	tg := notifier.NewTelegram(config.TelegramConfig{
		BotToken: job.TelegramBotToken,
		ChatID:   job.TelegramChatID,
	})
	return tg.Send(ctx, "✅ Telegram bağlantısı başarılı! Bot çalışıyor.")
}

func (s *jobServiceImpl) SubmitSMSCode(jobID string, code string) bool {
	return s.registry.Submit(jobID, code)
}

// ─── Durum ───────────────────────────────────────────────────────────────────

func (s *jobServiceImpl) GetJobStatus(ctx context.Context, id string) (*JobStatusResponse, error) {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := &JobStatusResponse{
		BookingJob: job,
		SMSWaiting: s.registry.IsWaiting(id),
	}

	s.mu.RLock()
	entry, ok := s.jobs[id]
	s.mu.RUnlock()

	if ok {
		resp.IsRunning = true
		status := entry.sched.GetStatus()
		resp.SchedulerMode = string(status.Mode)
		if !status.NextBurstAt.IsZero() {
			t := status.NextBurstAt
			resp.NextBurstAt = &t
			resp.NextBurstSlot = status.NextBurstSlot
		}
	}

	return resp, nil
}

// RestoreActiveJobs uygulama başladığında DB'deki aktif job'ları yeniden başlatır.
func (s *jobServiceImpl) RestoreActiveJobs(ctx context.Context) error {
	jobs, err := s.repo.ListByStatus(ctx, model.JobStatusActive)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		s.logger.Info("Aktif job yeniden başlatılıyor",
			zap.String("id", job.ID),
			zap.String("name", job.Name),
		)
		s.startJobInternal(job)
	}
	return nil
}

// ─── İç yardımcılar ──────────────────────────────────────────────────────────

func (s *jobServiceImpl) startJobInternal(job *model.BookingJob) {
	scraperCfg := jobToScraperConfig(job)
	creds := config.CredentialsConfig{TCNo: job.TCNo, Password: job.Password}

	br := browser.New(job.Headless, job.BrowserTimeoutSeconds, s.logger)

	var tgNotifier *notifier.TelegramNotifier
	if job.TelegramBotToken != "" && job.TelegramChatID != "" {
		tgNotifier = notifier.NewTelegram(config.TelegramConfig{
			BotToken:       job.TelegramBotToken,
			ChatID:         job.TelegramChatID,
			SuccessMessage: job.SuccessMessage,
		})
	}

	jobID := job.ID
	apiNotifier := NewAPINotifier(jobID, tgNotifier, s.registry)
	sc := scraper.New(scraperCfg, creds, br, apiNotifier, s.logger)

	runner := &jobRunner{
		sc:     sc,
		jobID:  jobID,
		repo:   s.repo,
		logger: s.logger,
	}

	sched := scheduler.New(scraperCfg, runner, s.logger)
	sched.Start()

	s.mu.Lock()
	s.jobs[jobID] = &jobEntry{sched: sched}
	s.mu.Unlock()
}

// jobRunner scraper.Run() çağrısını sararlar ve her run sonrasında DB'yi günceller.
type jobRunner struct {
	sc     *scraper.Scraper
	jobID  string
	repo   repository.JobRepository
	logger *zap.Logger
}

func (r *jobRunner) Run(ctx context.Context) (bool, error) {
	found, err := r.sc.Run(ctx)

	result := "slot bulunamadı"
	if err != nil {
		result = "HATA: " + err.Error()
	} else if found {
		result = "BULUNDU"
	}

	updateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = r.repo.IncrementRunStats(updateCtx, r.jobID, found, result)

	return found, err
}

func jobToScraperConfig(job *model.BookingJob) config.ScraperConfig {
	timeout := job.BrowserTimeoutSeconds
	if timeout == 0 {
		timeout = 60
	}
	return config.ScraperConfig{
		Facilities:            job.Facilities,
		SportType:             job.SportType,
		Courts:                job.Courts,
		TargetDates:           job.TargetDates,
		DesiredTimes:          job.DesiredTimes,
		OpeningOffsetHours:    job.OpeningOffsetHours,
		BurstBeforeSeconds:    job.BurstBeforeSeconds,
		BurstAfterSeconds:     job.BurstAfterSeconds,
		BurstIntervalSeconds:  job.BurstIntervalSeconds,
		PollIntervalSeconds:   job.PollIntervalSeconds,
		BrowserTimeoutSeconds: timeout,
		Headless:              job.Headless,
	}
}

func applyDefaults(job *model.BookingJob) {
	if job.OpeningOffsetHours == 0 {
		job.OpeningOffsetHours = 72
	}
	if job.BurstBeforeSeconds == 0 {
		job.BurstBeforeSeconds = 60
	}
	if job.BurstAfterSeconds == 0 {
		job.BurstAfterSeconds = 300
	}
	if job.BurstIntervalSeconds == 0 {
		job.BurstIntervalSeconds = 2
	}
	if job.PollIntervalSeconds == 0 {
		job.PollIntervalSeconds = 30
	}
	if job.BrowserTimeoutSeconds == 0 {
		job.BrowserTimeoutSeconds = 60
	}
}
