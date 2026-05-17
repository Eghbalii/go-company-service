package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is the top-level application configuration.
type Config struct {
	App      App
	Database Database
	JWT      JWT
	Kafka    Kafka
	Log      Log
}

// App holds HTTP server settings.
type App struct {
	Name  string `mapstructure:"APP_NAME"`
	Port  string `mapstructure:"APP_PORT"`
	Env   string `mapstructure:"APP_ENV"`
	Debug bool   `mapstructure:"APP_DEBUG"`
}

// Database holds PostgreSQL connection settings.
type Database struct {
	Host            string        `mapstructure:"DB_HOST"`
	Port            string        `mapstructure:"DB_PORT"`
	User            string        `mapstructure:"DB_USER"`
	Password        string        `mapstructure:"DB_PASSWORD"`
	Name            string        `mapstructure:"DB_NAME"`
	SSLMode         string        `mapstructure:"DB_SSLMODE"`
	MaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`
}

// DSN returns the PostgreSQL connection string.
func (d Database) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// JWT holds token signing settings.
type JWT struct {
	Secret string        `mapstructure:"JWT_SECRET"`
	Expiry time.Duration `mapstructure:"JWT_EXPIRY"`
}

// Kafka holds broker and topic configuration.
type Kafka struct {
	Brokers []string `mapstructure:"KAFKA_BROKERS"`
	Topic   string   `mapstructure:"KAFKA_TOPIC_COMPANIES"`
	Async   bool     `mapstructure:"KAFKA_WRITER_ASYNC"`
}

// Log holds structured logging settings.
type Log struct {
	Level  string `mapstructure:"LOG_LEVEL"`
	Format string `mapstructure:"LOG_FORMAT"`
}

// Load reads configuration from environment variables and an optional .env file.
// Environment variables always take precedence over .env values.
func Load() (*Config, error) {
	v := viper.New()

	setDefaults(v)

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig() // .env is optional

	v.AutomaticEnv()

	// Comma-separated KAFKA_BROKERS support (e.g. "broker1:9092,broker2:9092")
	if brokers := v.GetString("KAFKA_BROKERS"); brokers != "" && !strings.Contains(brokers, "[") {
		v.Set("KAFKA_BROKERS", strings.Split(brokers, ","))
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET must be set")
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_NAME", "company-service")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_DEBUG", false)

	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_PASSWORD", "postgres")
	v.SetDefault("DB_NAME", "company_db")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 5)
	v.SetDefault("DB_CONN_MAX_LIFETIME", "5m")

	v.SetDefault("JWT_SECRET", "")
	v.SetDefault("JWT_EXPIRY", "24h")

	v.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	v.SetDefault("KAFKA_TOPIC_COMPANIES", "company-events")
	v.SetDefault("KAFKA_WRITER_ASYNC", true)

	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")
}
