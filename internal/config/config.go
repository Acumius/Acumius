package config

import (
	"os"
	"time"
)

const (
	envHTTPAddr          = "ACUMIUS_HTTP_ADDR"
	envReadHeaderTimeout = "ACUMIUS_HTTP_READ_HEADER_TIMEOUT"
	envReadTimeout       = "ACUMIUS_HTTP_READ_TIMEOUT"
	envWriteTimeout      = "ACUMIUS_HTTP_WRITE_TIMEOUT"
	envIdleTimeout       = "ACUMIUS_HTTP_IDLE_TIMEOUT"
	envShutdownTimeout   = "ACUMIUS_SHUTDOWN_TIMEOUT"
)

// Config is the runtime configuration for the Acumius service.
type Config struct {
	HTTPAddr          string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

// Load reads configuration from environment variables and applies defaults.
func Load() Config {
	return Config{
		HTTPAddr:          getEnv(envHTTPAddr, ":8080"),
		ReadHeaderTimeout: parseDurationEnv(envReadHeaderTimeout, "5s"),
		ReadTimeout:       parseDurationEnv(envReadTimeout, "10s"),
		WriteTimeout:      parseDurationEnv(envWriteTimeout, "15s"),
		IdleTimeout:       parseDurationEnv(envIdleTimeout, "60s"),
		ShutdownTimeout:   parseDurationEnv(envShutdownTimeout, "10s"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return fallback
}

func parseDurationEnv(key, fallback string) time.Duration {
	value := getEnv(key, fallback)
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return mustParseDuration(fallback)
	}

	return parsed
}

func mustParseDuration(value string) time.Duration {
	parsed, err := time.ParseDuration(value)
	if err != nil {
		panic(err)
	}

	return parsed
}
