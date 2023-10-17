package app

import (
	"context"
	"fmt"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/gophermart/internal/controllers/handlers"
	"github.com/pavlegich/gophermart/internal/controllers/middlewares"
	"github.com/pavlegich/gophermart/internal/infra/config"
	"github.com/pavlegich/gophermart/internal/infra/database"
	"github.com/pavlegich/gophermart/internal/infra/logger"
	"go.uber.org/zap"
)

// Run инициализирует основные компоненты и запускает сервер
func Run() error {
	// Контекст
	ctx := context.Background()

	// Логгер
	if err := logger.Init(ctx, "Info"); err != nil {
		return fmt.Errorf("Run: logger initialization failed %w", err)
	}
	defer logger.Log.Sync()

	// Конфиг
	cfg, err := config.ParseFlags(ctx)
	if err != nil {
		return fmt.Errorf("Run: parse flags failed %w", err)
	}

	fmt.Printf("=====\nCONFIGURATION: %s\n=====\n", cfg)

	// База данных
	db, err := database.Init(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("Run: database initialization failed %w", err)
	}
	defer db.Close()

	// Контроллер
	server := handlers.NewController(db, cfg)
	serverRouter, err := server.BuildRoute(ctx)
	if err != nil {
		return fmt.Errorf("Run: build server router failed %w", err)
	}

	// Роутер
	r := chi.NewRouter()
	r.Use(middlewares.Recovery)
	r.Mount("/", serverRouter)

	logger.Log.Info("Running server", zap.String("address", cfg.Address))

	return http.ListenAndServe(cfg.Address, r)
}
