package config

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/pavlegich/gophermart/internal/infra/hash"
)

// Config хранит значения флагов, ключей или переменных окружения
type Config struct {
	Address   string `env:"RUN_ADDRESS"`
	Database  string `env:"DATABASE_URI"`
	Accrual   string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	Update    time.Duration
	RateLimit int
	JWT       *hash.JWT
}

// ParseFlags обрабатывает значения флагов и переменных окружения
func ParseFlags(ctx context.Context) (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Gophermart service running host:port")
	flag.StringVar(&cfg.Database, "d", "postgresql://localhost:5432/gophermart", "URI (DSN) to database")
	flag.StringVar(&cfg.Accrual, "r", "http://localhost:8088", "Accrual service host:port")

	cfg.Update = 5 * time.Second
	cfg.RateLimit = 1

	// Создание ключей для JWT
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return cfg, fmt.Errorf("ParseFlags: generate private key failed")
	}
	tokenExp := 3 * time.Hour
	cfg.JWT = hash.NewJWT(privateKey, &privateKey.PublicKey, tokenExp)

	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		return cfg, fmt.Errorf("ParseFlags: wrong environment values %w", err)
	}

	return cfg, nil
}
