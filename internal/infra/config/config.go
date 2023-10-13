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
	address  string `env:"RUN_ADDRESS"`
	database string `env:"DATABASE_URI"`
	accrual  string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func (c *Config) GetAddress() string {
	return c.address
}

func (c *Config) GetDBuri() string {
	return c.database
}

func (c *Config) GetAccrualAddr() string {
	return c.accrual
}

// ParseFlags обрабатывает значения флагов и переменных окружения
func ParseFlags(ctx context.Context) (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.address, "a", "localhost:8080", "Gophermart service running host:port")
	flag.StringVar(&cfg.database, "d", "postgresql://localhost:5432/gophermart", "URI (DSN) to database")
	flag.StringVar(&cfg.accrual, "r", "", "Accrual service host:port")

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
	if err := checkAddress(cfg.GetAddress()); err != nil {
		return fmt.Errorf("CheckConfig: check accrual failed %w", err)
	}

	if cfg.GetDBuri() == "" {
		return fmt.Errorf("CheckConfig: database address required")
	}

	if cfg.GetAccrualAddr() != "" {
		if err := checkAddress(cfg.GetAccrualAddr()); err != nil {
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
