-- Appointment Scrapper - İlk migration
-- booking_jobs tablosunu oluşturur.
-- NOT: Uygulama başladığında bu migration otomatik çalışır (CREATE TABLE IF NOT EXISTS).
-- Bu dosya manuel psql ile de çalıştırılabilir:
--   psql "host=... dbname=..." -f migrations/001_create_booking_jobs.sql

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
