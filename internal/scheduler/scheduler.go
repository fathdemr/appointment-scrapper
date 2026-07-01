package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"appointment-scrapper/config"
	"appointment-scrapper/internal/scraper"
)

// SlotTarget bir hedef randevu slotunu ve onun açılış zamanını tanımlar.
type SlotTarget struct {
	SlotTime time.Time // örn: 04.07.2026 21:00
	OpensAt  time.Time // SlotTime - opening_offset_hours
}

type Mode string

const (
	ModeIdle   Mode = "idle"
	ModeNormal Mode = "normal"
	ModeBurst  Mode = "burst"
)

type Status struct {
	Running       bool      `json:"running"`
	Mode          Mode      `json:"mode"`
	LastRun       time.Time `json:"last_run"`
	LastResult    string    `json:"last_result"`
	TotalRuns     int       `json:"total_runs"`
	TotalFound    int       `json:"total_found"`
	NextBurstAt   time.Time `json:"next_burst_at,omitempty"`
	NextBurstSlot string    `json:"next_burst_slot,omitempty"`
}

type Scheduler struct {
	cfg     config.ScraperConfig
	scraper *scraper.Scraper
	logger  *zap.Logger

	stopCh chan struct{}
	mu     sync.Mutex
	status Status
}

func New(cfg config.ScraperConfig, sc *scraper.Scraper, logger *zap.Logger) *Scheduler {
	return &Scheduler{
		cfg:     cfg,
		scraper: sc,
		logger:  logger,
		stopCh:  make(chan struct{}),
		status:  Status{Mode: ModeIdle},
	}
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	s.status.Running = true
	s.mu.Unlock()

	s.logUpcomingBursts()
	go s.loop()

	s.logger.Info("Zamanlayıcı başlatıldı",
		zap.Int("opening_offset_hours", s.cfg.OpeningOffsetHours),
		zap.Int("burst_before_seconds", s.cfg.BurstBeforeSeconds),
		zap.Int("burst_after_seconds", s.cfg.BurstAfterSeconds),
		zap.Strings("target_dates", s.cfg.TargetDates),
		zap.Strings("desired_times", s.cfg.DesiredTimes),
	)
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.mu.Lock()
	s.status.Running = false
	s.status.Mode = ModeIdle
	s.mu.Unlock()
	s.logger.Info("Zamanlayıcı durduruldu")
}

func (s *Scheduler) GetStatus() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.status
	next, label := s.nextBurst(time.Now())
	if !next.IsZero() {
		st.NextBurstAt = next
		st.NextBurstSlot = label
	}
	return st
}

func (s *Scheduler) RunNow() {
	go s.tick(time.Now())
}

func (s *Scheduler) loop() {
	for {
		now := time.Now()
		interval, mode := s.selectInterval(now)

		s.mu.Lock()
		s.status.Mode = mode
		s.mu.Unlock()

		if mode == ModeBurst {
			s.logger.Info("BURST MODU aktif", zap.Duration("interval", interval))
		}

		select {
		case <-s.stopCh:
			return
		case <-time.After(interval):
			s.tick(time.Now())
		}
	}
}

func (s *Scheduler) selectInterval(now time.Time) (time.Duration, Mode) {
	if s.isInBurstWindow(now) {
		return time.Duration(s.cfg.BurstIntervalSeconds) * time.Second, ModeBurst
	}
	return time.Duration(s.cfg.PollIntervalSeconds) * time.Second, ModeNormal
}

func (s *Scheduler) isInBurstWindow(now time.Time) bool {
	before := time.Duration(s.cfg.BurstBeforeSeconds) * time.Second
	after := time.Duration(s.cfg.BurstAfterSeconds) * time.Second
	for _, t := range s.buildTargets(now) {
		if now.After(t.OpensAt.Add(-before)) && now.Before(t.OpensAt.Add(after)) {
			return true
		}
	}
	return false
}

func (s *Scheduler) tick(now time.Time) {
	s.mu.Lock()
	s.status.LastRun = now
	s.status.TotalRuns++
	mode := s.status.Mode
	s.mu.Unlock()

	s.logger.Info("Scraper çalıştırılıyor", zap.String("mode", string(mode)))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	found, err := s.scraper.Run(ctx)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err != nil {
		s.status.LastResult = "HATA: " + err.Error()
		s.logger.Error("Scraper hatası", zap.Error(err))
		return
	}
	if found {
		s.status.TotalFound++
		s.status.LastResult = "BULUNDU"
	} else {
		s.status.LastResult = "Slot bulunamadı"
	}
}

// buildTargets config'deki target_dates × desired_times kombinasyonlarından
// SlotTarget listesi üretir. Geçmiş burst pencereleri dahil edilmez.
func (s *Scheduler) buildTargets(from time.Time) []SlotTarget {
	offset := time.Duration(s.cfg.OpeningOffsetHours) * time.Hour
	burstAfter := time.Duration(s.cfg.BurstAfterSeconds) * time.Second
	var targets []SlotTarget

	for _, dateStr := range s.cfg.TargetDates {
		date, err := time.ParseInLocation("02.01.2006", dateStr, from.Location())
		if err != nil {
			s.logger.Warn("Tarih parse edilemedi", zap.String("date", dateStr), zap.Error(err))
			continue
		}

		for _, timeStr := range s.cfg.DesiredTimes {
			slotTime := parseOnDay(date, timeStr)
			if slotTime.IsZero() {
				continue
			}
			opensAt := slotTime.Add(-offset)

			// Burst penceresi tamamen geçtiyse atla
			if opensAt.Add(burstAfter).Before(from) {
				continue
			}
			targets = append(targets, SlotTarget{
				SlotTime: slotTime,
				OpensAt:  opensAt,
			})
		}
	}
	return targets
}

func (s *Scheduler) nextBurst(now time.Time) (time.Time, string) {
	before := time.Duration(s.cfg.BurstBeforeSeconds) * time.Second
	var earliest time.Time
	var label string

	for _, t := range s.buildTargets(now) {
		burstStart := t.OpensAt.Add(-before)
		if burstStart.After(now) && (earliest.IsZero() || burstStart.Before(earliest)) {
			earliest = burstStart
			label = t.SlotTime.Format("02.01.2006 15:04")
		}
	}
	return earliest, label
}

func (s *Scheduler) logUpcomingBursts() {
	now := time.Now()
	before := time.Duration(s.cfg.BurstBeforeSeconds) * time.Second
	after := time.Duration(s.cfg.BurstAfterSeconds) * time.Second

	for _, t := range s.buildTargets(now) {
		s.logger.Info("Burst penceresi",
			zap.String("hedef", t.SlotTime.Format("02.01.2006 15:04 (Mon)")),
			zap.String("acilis", t.OpensAt.Format("02.01.2006 15:04:05 (Mon)")),
			zap.String("burst_baslangic", t.OpensAt.Add(-before).Format("02.01 15:04:05")),
			zap.String("burst_bitis", t.OpensAt.Add(after).Format("02.01 15:04:05")),
		)
	}
}

func parseOnDay(day time.Time, hhmm string) time.Time {
	var h, m int
	if _, err := fmt.Sscanf(hhmm, "%d:%d", &h, &m); err != nil {
		return time.Time{}
	}
	return time.Date(day.Year(), day.Month(), day.Day(), h, m, 0, 0, day.Location())
}
