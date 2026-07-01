package main

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"appointment-scrapper/config"
	"appointment-scrapper/internal/api"
	"appointment-scrapper/internal/browser"
	"appointment-scrapper/internal/notifier"
	"appointment-scrapper/internal/scheduler"
	"appointment-scrapper/internal/scraper"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("Config yüklenemedi: " + err.Error())
	}

	logger := buildLogger(cfg.App.LogLevel)
	defer logger.Sync()

	tg := notifier.NewTelegram(cfg.Telegram)
	br := browser.New(cfg.Scraper, logger)
	sc := scraper.New(cfg.Scraper, cfg.Credentials, br, tg, logger)

	sched := scheduler.New(cfg.Scraper, sc, logger)
	sched.Start()
	defer sched.Stop()

	srv := api.New(cfg.App, sched, logger)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			logger.Error("Sunucu hatası", zap.Error(err))
		}
	}()

	logger.Info("Uygulama çalışıyor", zap.Int("port", cfg.App.Port))
	<-quit
	logger.Info("Kapatılıyor...")
}

func buildLogger(level string) *zap.Logger {
	var lvl zapcore.Level
	switch level {
	case "debug":
		lvl = zapcore.DebugLevel
	case "warn":
		lvl = zapcore.WarnLevel
	case "error":
		lvl = zapcore.ErrorLevel
	default:
		lvl = zapcore.InfoLevel
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	logger, _ := cfg.Build()
	return logger
}
