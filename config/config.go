package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App App `mapstructure:"app"`
	DB  DB  `mapstructure:"db"`

	// Aşağıdaki alanlar cmd/inspect aracı tarafından config.yaml'dan okunur.
	// Üretim servisi bu değerleri DB'deki job kayıtlarından alır.
	Credentials CredentialsConfig `mapstructure:"credentials"`
	Scraper     ScraperConfig     `mapstructure:"scraper"`
	Telegram    TelegramConfig    `mapstructure:"telegram"`
}

type App struct {
	Port     int    `mapstructure:"port"`
	LogLevel string `mapstructure:"log_level"`
	APIKey   string `mapstructure:"api_key"`
}

type DB struct {
	Host     string `mapstructure:"host"`
	DBName   string `mapstructure:"db_name"`
	UserName string `mapstructure:"user_name"`
	Password string `mapstructure:"password"`
	Port     string `mapstructure:"port"`
}

// DSN PostgreSQL bağlantı dizesini döner.
func (d DB) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		d.Host, d.Port, d.UserName, d.Password, d.DBName,
	)
}

// ─── Sadece iç kullanım için tip tanımları (cmd/inspect + scraper/scheduler) ─

type CredentialsConfig struct {
	TCNo     string `mapstructure:"tc_no"`
	Password string `mapstructure:"password"`
}

type ScraperConfig struct {
	Facilities  []string `mapstructure:"facility"`
	SportType   string   `mapstructure:"sport_type"`
	Courts      []string `mapstructure:"courts"`
	TargetDates []string `mapstructure:"target_dates"`

	DesiredTimes []string `mapstructure:"desired_times"`

	OpeningOffsetHours   int `mapstructure:"opening_offset_hours"`
	BurstBeforeSeconds   int `mapstructure:"burst_before_seconds"`
	BurstAfterSeconds    int `mapstructure:"burst_after_seconds"`
	BurstIntervalSeconds int `mapstructure:"burst_interval_seconds"`
	PollIntervalSeconds  int `mapstructure:"poll_interval_seconds"`

	BrowserTimeoutSeconds int  `mapstructure:"browser_timeout_seconds"`
	Headless              bool `mapstructure:"headless"`
}

type TelegramConfig struct {
	BotToken       string `mapstructure:"bot_token"`
	ChatID         string `mapstructure:"chat_id"`
	SuccessMessage string `mapstructure:"success_message"`
}

// AppConfig eski API uyumluluğu için takma ad
type AppConfig = App

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.MergeInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
