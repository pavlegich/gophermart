package config

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
)

// Config хранит значения флагов или переменных окружения
type Config struct {
	Address   string `env:"RUN_ADDRESS"`
	Database  string `env:"DATABASE_URI"`
	Accrual   string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	Update    time.Duration
	RateLimit int
}

// ParseFlags обрабатывает значения флагов и переменных окружения
func ParseFlags(ctx context.Context) (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Gophermart service running host:port")
	flag.StringVar(&cfg.Database, "d", "postgresql://localhost:5432/gophermart", "URI (DSN) to database")
	flag.StringVar(&cfg.Accrual, "r", "", "Accrual service host:port")

	cfg.Update = 2 * time.Second
	cfg.RateLimit = 1

	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		return cfg, fmt.Errorf("ParseFlags: wrong environment values %w", err)
	}

	return cfg, nil
}
