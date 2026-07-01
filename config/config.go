package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App         AppConfig         `mapstructure:"app"`
	Credentials CredentialsConfig `mapstructure:"credentials"`
	Scraper     ScraperConfig     `mapstructure:"scraper"`
	Telegram    TelegramConfig    `mapstructure:"telegram"`
}

type AppConfig struct {
	Port     int    `mapstructure:"port"`
	LogLevel string `mapstructure:"log_level"`
}

type CredentialsConfig struct {
	TCNo     string `mapstructure:"tc_no"`
	Password string `mapstructure:"password"`
}

type ScraperConfig struct {
	Facilities  []string `mapstructure:"facility"`
	SportType   string   `mapstructure:"sport_type"`
	Courts      []string `mapstructure:"courts"`
	TargetDates []string `mapstructure:"target_dates"` // DD.MM.YYYY formatında

	DesiredTimes []string `mapstructure:"desired_times"`

	// 72 saat kuralı: slot N saat öncesinde açılır
	OpeningOffsetHours int `mapstructure:"opening_offset_hours"`

	// Açılış anı etrafında burst polling
	BurstBeforeSeconds   int `mapstructure:"burst_before_seconds"`
	BurstAfterSeconds    int `mapstructure:"burst_after_seconds"`
	BurstIntervalSeconds int `mapstructure:"burst_interval_seconds"`

	// Normal polling (iptal yakalamak için)
	PollIntervalSeconds int `mapstructure:"poll_interval_seconds"`

	BrowserTimeoutSeconds int  `mapstructure:"browser_timeout_seconds"`
	Headless              bool `mapstructure:"headless"`
}

type TelegramConfig struct {
	BotToken       string `mapstructure:"bot_token"`
	ChatID         string `mapstructure:"chat_id"`
	SuccessMessage string `mapstructure:"success_message"`
}

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

	// .env dosyası varsa hassas değerleri override et
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.MergeInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
