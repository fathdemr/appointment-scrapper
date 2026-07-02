// cmd/migrate — GORM AutoMigrate ile katalog tablolarını yönetir.
// Kullanım: go run ./cmd/migrate
package main

import (
	"fmt"
	"log"

	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"appointment-scrapper/config"
	"appointment-scrapper/model"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config yüklenemedi: %v", err)
	}

	db, err := gorm.Open(gormpg.Open(cfg.DB.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("DB bağlantısı kurulamadı: %v", err)
	}

	fmt.Println("Migration başlıyor...")

	if err := db.AutoMigrate(
		&model.SportType{},
		&model.Facility{},
		&model.Court{},
	); err != nil {
		log.Fatalf("Migration başarısız: %v", err)
	}

	fmt.Println("Migration tamamlandı.")
	fmt.Println("  ✓ sport_types")
	fmt.Println("  ✓ facilities")
	fmt.Println("  ✓ courts")
}
