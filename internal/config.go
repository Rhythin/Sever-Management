package internal

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all environment configuration for the service
// Use `envconfig` struct tags for env var mapping and defaults

type Config struct {
	Env            string        `envconfig:"ENV" default:"development"`
	HTTPPort       int           `envconfig:"HTTP_PORT" default:"8080"`
	DBHost         string        `envconfig:"DB_HOST" default:"localhost"`
	DBPort         int           `envconfig:"DB_PORT" default:"5432"`
	DBUser         string        `envconfig:"DB_USER" default:"postgres"`
	DBPassword     string        `envconfig:"DB_PASSWORD" default:"password"`
	DBName         string        `envconfig:"DB_NAME" default:"servermgmt"`
	DBSSLMode      string        `envconfig:"DB_SSLMODE" default:"disable"`
	BillingRate    float64       `envconfig:"BILLING_RATE" default:"0.01"` // $/hr
	IdleTimeout    time.Duration `envconfig:"IDLE_TIMEOUT" default:"30m"`
	IPCIDR         string        `envconfig:"IP_CIDR" default:"192.168.0.0/16"`
	LogLevel       string        `envconfig:"LOG_LEVEL" default:"info"`
	MetricsPort    int           `envconfig:"METRICS_PORT" default:"9090"`
	RequestTimeout time.Duration `envconfig:"REQUEST_TIMEOUT" default:"10s"`
}

// LoadConfig loads config from environment variables
func LoadConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &cfg, nil
}
