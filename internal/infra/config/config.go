package config

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
)

// Config хранит значения флагов или переменных окружения
type Config struct {
	Address  string `env:"RUN_ADDRESS"`
	Database string `env:"DATABASE_URI"`
	Accrual  string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

// ParseFlags обрабатывает значения флагов и переменных окружения
func ParseFlags(ctx context.Context) (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Gophermart service running host:port")
	flag.StringVar(&cfg.Database, "d", "postgresql://localhost:5432/gophermart", "URI (DSN) to database")
	flag.StringVar(&cfg.Accrual, "r", "", "Accrual service host:port")

	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("ParseFlags: wrong environment values %w", err)
	}

	if err := checkConfig(cfg); err != nil {
		return nil, fmt.Errorf("ParseFlags: check config failed %w", err)
	}

	return cfg, nil
}

// checkConfig проверяет корректность полученных данных конфигурации
func checkConfig(cfg *Config) error {
	if err := checkAddress(cfg.Address); err != nil {
		return fmt.Errorf("CheckConfig: check accrual failed %w", err)
	}

	if cfg.Database == "" {
		return fmt.Errorf("CheckConfig: database address required")
	}

	if cfg.Accrual != "" {
		if err := checkAddress(cfg.Accrual); err != nil {
			return fmt.Errorf("CheckConfig: check accrual failed %w", err)
		}
	}

	return nil
}

// checkAddress проверяет корректность адреса
func checkAddress(addr string) error {
	values := strings.Split(addr, ":")

	if len(strings.Split(addr, ":")) != 2 {
		return fmt.Errorf("checkAddress: address '%s' not in a form host:port", addr)
	}
	_, err := strconv.Atoi(values[1])
	if err != nil {
		return fmt.Errorf("checkAddress: convert port '%s' failed %w", values[1], err)
	}

	return nil
}
