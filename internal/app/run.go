package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/gophermart/internal/controllers/handlers"
	"github.com/pavlegich/gophermart/internal/controllers/middlewares"
	"github.com/pavlegich/gophermart/internal/infra/config"
	"github.com/pavlegich/gophermart/internal/infra/logger"
	"go.uber.org/zap"
)

// Run инициализирует основные компоненты и запускает сервер
func Run() error {
	// Контекст
	ctx := context.Background()

	// Инициализация логгера
	if err := logger.Initialize(ctx, "Info"); err != nil {
		return fmt.Errorf("Run: logger initialization failed %w", err)
	}
	defer logger.Log.Sync()

	// Флаги
	cfg, err := config.ParseFlags(ctx)
	if err != nil {
		return fmt.Errorf("Run: parse flags failed %w", err)
	}

	// Инициализация контроллера
	controller := handlers.NewController(cfg)

	// Роутер
	r := chi.NewRouter()
	r.Use(middlewares.Recovery)
	r.Mount("/", controller.Route(ctx))

	logger.Log.Info("Running server", zap.String("address", controller.GetAddress()))

	return http.ListenAndServe(controller.GetAddress(), r)
}
