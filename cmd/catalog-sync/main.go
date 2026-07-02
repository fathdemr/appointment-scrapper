// cmd/catalog-sync — spor.istanbul'dan katalog verisini çeker, DB'ye yazar.
// Kullanım: go run ./cmd/catalog-sync
// Gereksinim: config.yaml içinde db ve credentials alanları dolu olmalı.
package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"appointment-scrapper/config"
	"appointment-scrapper/internal/catalog"
	catalogrepo "appointment-scrapper/repository/catalog"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config yüklenemedi: %v", err)
	}
	if cfg.Credentials.TCNo == "" || cfg.Credentials.Password == "" {
		log.Fatal("config.yaml içinde credentials.tc_no ve credentials.password dolu olmalı")
	}

	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	zapCfg.EncoderConfig.TimeKey = "time"
	zapCfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	logger, _ := zapCfg.Build()
	defer logger.Sync()

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DB.DSN())
	if err != nil {
		logger.Fatal("DB bağlantısı kurulamadı", zap.Error(err))
	}
	defer pool.Close()

	cRepo := catalogrepo.New(pool)
	s := catalog.NewScraper(cRepo, logger)

	syncCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	logger.Info("Katalog sync başlıyor...",
		zap.String("tc_no", cfg.Credentials.TCNo[:3]+"****"))

	if err := s.Sync(syncCtx, cfg.Credentials.TCNo, cfg.Credentials.Password); err != nil {
		logger.Fatal("Sync başarısız", zap.Error(err))
	}
	logger.Info("Katalog sync tamamlandı")
}
