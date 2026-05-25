package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config is the runtime configuration for the Acumius service.
type Config struct {
	ServerPort  string `env:"ACUMIUS_SERVER_PORT" envDefault:"8080"`
	DatabaseURL string `env:"ACUMIUS_DATABASE_URL" envDefault:"postgres://acumius:acumius@localhost:5432/acumius?sslmode=disable"`
	ValkeyURL   string `env:"ACUMIUS_VALKEY_URL" envDefault:"localhost:6379"`
	LogLevel    string `env:"ACUMIUS_LOG_LEVEL" envDefault:"info"`
	Environment string `env:"ACUMIUS_ENV" envDefault:"development"`

	ReadHeaderTimeout time.Duration `env:"ACUMIUS_HTTP_READ_HEADER_TIMEOUT" envDefault:"5s"`
	ReadTimeout       time.Duration `env:"ACUMIUS_HTTP_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout      time.Duration `env:"ACUMIUS_HTTP_WRITE_TIMEOUT" envDefault:"15s"`
	IdleTimeout       time.Duration `env:"ACUMIUS_HTTP_IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout   time.Duration `env:"ACUMIUS_SHUTDOWN_TIMEOUT" envDefault:"10s"`
}

// Load reads configuration from environment variables and applies defaults.
func Load() Config {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	return cfg
}
