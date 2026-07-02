package pgxrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"appointment-scrapper/model"
	"appointment-scrapper/repository"
)

const selectCols = `
	id, name, status, tc_no, password, sport_type,
	facilities, courts, target_dates, desired_times,
	telegram_bot_token, telegram_chat_id, success_message,
	opening_offset_hours, burst_before_seconds, burst_after_seconds,
	burst_interval_seconds, poll_interval_seconds,
	headless, browser_timeout_seconds,
	last_result, total_runs, total_found, last_run_at,
	created_at, updated_at
`

const createTableSQL = `
CREATE TABLE IF NOT EXISTS booking_jobs (
	id                      TEXT PRIMARY KEY,
	name                    TEXT NOT NULL,
	status                  TEXT NOT NULL DEFAULT 'pending',

	tc_no                   TEXT NOT NULL,
	password                TEXT NOT NULL,

	sport_type              TEXT NOT NULL DEFAULT '',
	facilities              TEXT[] NOT NULL DEFAULT '{}',
	courts                  TEXT[] NOT NULL DEFAULT '{}',
	target_dates            TEXT[] NOT NULL DEFAULT '{}',
	desired_times           TEXT[] NOT NULL DEFAULT '{}',

	telegram_bot_token      TEXT NOT NULL DEFAULT '',
	telegram_chat_id        TEXT NOT NULL DEFAULT '',
	success_message         TEXT NOT NULL DEFAULT '',

	opening_offset_hours    INT NOT NULL DEFAULT 72,
	burst_before_seconds    INT NOT NULL DEFAULT 60,
	burst_after_seconds     INT NOT NULL DEFAULT 300,
	burst_interval_seconds  INT NOT NULL DEFAULT 2,
	poll_interval_seconds   INT NOT NULL DEFAULT 30,
	headless                BOOLEAN NOT NULL DEFAULT false,
	browser_timeout_seconds INT NOT NULL DEFAULT 60,

	last_result             TEXT NOT NULL DEFAULT '',
	total_runs              INT NOT NULL DEFAULT 0,
	total_found             INT NOT NULL DEFAULT 0,
	last_run_at             TIMESTAMPTZ,

	created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_booking_jobs_status ON booking_jobs(status);
`

type JobPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewJobRepository(dsn string) (*JobPostgresRepository, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}

	repo := &JobPostgresRepository{pool: pool}
	if err := repo.migrate(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migration: %w", err)
	}
	return repo, nil
}

func (r *JobPostgresRepository) migrate(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, createTableSQL)
	return err
}

func (r *JobPostgresRepository) Create(ctx context.Context, job *model.BookingJob) error {
	job.ID = uuid.New().String()
	now := time.Now()
	job.CreatedAt = now
	job.UpdatedAt = now
	if job.Status == "" {
		job.Status = model.JobStatusPending
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO booking_jobs (
			id, name, status, tc_no, password, sport_type,
			facilities, courts, target_dates, desired_times,
			telegram_bot_token, telegram_chat_id, success_message,
			opening_offset_hours, burst_before_seconds, burst_after_seconds,
			burst_interval_seconds, poll_interval_seconds,
			headless, browser_timeout_seconds,
			last_result, total_runs, total_found,
			created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,
			$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
			$21,$22,$23,$24,$25
		)`,
		job.ID, job.Name, string(job.Status), job.TCNo, job.Password,
		job.SportType, job.Facilities, job.Courts, job.TargetDates, job.DesiredTimes,
		job.TelegramBotToken, job.TelegramChatID, job.SuccessMessage,
		job.OpeningOffsetHours, job.BurstBeforeSeconds, job.BurstAfterSeconds,
		job.BurstIntervalSeconds, job.PollIntervalSeconds,
		job.Headless, job.BrowserTimeoutSeconds,
		job.LastResult, job.TotalRuns, job.TotalFound,
		job.CreatedAt, job.UpdatedAt,
	)
	return err
}

func (r *JobPostgresRepository) GetByID(ctx context.Context, id string) (*model.BookingJob, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+selectCols+` FROM booking_jobs WHERE id=$1`, id)
	job, err := scanJob(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return job, nil
}

func (r *JobPostgresRepository) List(ctx context.Context) ([]*model.BookingJob, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+selectCols+` FROM booking_jobs ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectJobs(rows)
}

func (r *JobPostgresRepository) ListByStatus(ctx context.Context, status model.JobStatus) ([]*model.BookingJob, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+selectCols+` FROM booking_jobs WHERE status=$1 ORDER BY created_at DESC`,
		string(status))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectJobs(rows)
}

func (r *JobPostgresRepository) Update(ctx context.Context, job *model.BookingJob) error {
	job.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE booking_jobs SET
			name=$1, status=$2, tc_no=$3, password=$4, sport_type=$5,
			facilities=$6, courts=$7, target_dates=$8, desired_times=$9,
			telegram_bot_token=$10, telegram_chat_id=$11, success_message=$12,
			opening_offset_hours=$13, burst_before_seconds=$14, burst_after_seconds=$15,
			burst_interval_seconds=$16, poll_interval_seconds=$17,
			headless=$18, browser_timeout_seconds=$19,
			updated_at=$20
		WHERE id=$21`,
		job.Name, string(job.Status), job.TCNo, job.Password, job.SportType,
		job.Facilities, job.Courts, job.TargetDates, job.DesiredTimes,
		job.TelegramBotToken, job.TelegramChatID, job.SuccessMessage,
		job.OpeningOffsetHours, job.BurstBeforeSeconds, job.BurstAfterSeconds,
		job.BurstIntervalSeconds, job.PollIntervalSeconds,
		job.Headless, job.BrowserTimeoutSeconds,
		job.UpdatedAt, job.ID,
	)
	return err
}

func (r *JobPostgresRepository) Delete(ctx context.Context, id string) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM booking_jobs WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *JobPostgresRepository) UpdateStatus(ctx context.Context, id string, status model.JobStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE booking_jobs SET status=$1, updated_at=$2 WHERE id=$3`,
		string(status), time.Now(), id,
	)
	return err
}

func (r *JobPostgresRepository) IncrementRunStats(ctx context.Context, id string, found bool, result string) error {
	foundDelta := 0
	if found {
		foundDelta = 1
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE booking_jobs SET
			total_runs  = total_runs + 1,
			total_found = total_found + $1,
			last_result = $2,
			last_run_at = $3,
			updated_at  = $3
		WHERE id=$4`,
		foundDelta, result, time.Now(), id,
	)
	return err
}

// ─── İç yardımcılar ──────────────────────────────────────────────────────────

type rowScanner interface {
	Scan(dest ...any) error
}

func scanJob(s rowScanner) (*model.BookingJob, error) {
	var (
		status    string
		lastRunAt pgtype.Timestamptz
	)
	job := &model.BookingJob{}

	err := s.Scan(
		&job.ID,
		&job.Name,
		&status,
		&job.TCNo,
		&job.Password,
		&job.SportType,
		&job.Facilities,
		&job.Courts,
		&job.TargetDates,
		&job.DesiredTimes,
		&job.TelegramBotToken,
		&job.TelegramChatID,
		&job.SuccessMessage,
		&job.OpeningOffsetHours,
		&job.BurstBeforeSeconds,
		&job.BurstAfterSeconds,
		&job.BurstIntervalSeconds,
		&job.PollIntervalSeconds,
		&job.Headless,
		&job.BrowserTimeoutSeconds,
		&job.LastResult,
		&job.TotalRuns,
		&job.TotalFound,
		&lastRunAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	job.Status = model.JobStatus(status)
	if lastRunAt.Valid {
		t := lastRunAt.Time
		job.LastRunAt = &t
	}
	return job, nil
}

func collectJobs(rows pgx.Rows) ([]*model.BookingJob, error) {
	var jobs []*model.BookingJob
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if jobs == nil {
		jobs = []*model.BookingJob{}
	}
	return jobs, nil
}
