package app

import (
	"context"
	"fmt"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/gophermart/internal/controllers/handlers"
	"github.com/pavlegich/gophermart/internal/controllers/middlewares"
	"github.com/pavlegich/gophermart/internal/domains/user"
	userHandler "github.com/pavlegich/gophermart/internal/domains/user/controllers/http"
	userRepo "github.com/pavlegich/gophermart/internal/domains/user/repository"
	"github.com/pavlegich/gophermart/internal/infra/config"
	"github.com/pavlegich/gophermart/internal/infra/db"
	"github.com/pavlegich/gophermart/internal/infra/logger"
	"go.uber.org/zap"
)

// Run инициализирует основные компоненты и запускает сервер
func Run() error {
	// Контекст
	ctx := context.Background()

	// Логгер
	if err := logger.Initialize(ctx, "Info"); err != nil {
		return fmt.Errorf("Run: logger initialization failed %w", err)
	}
	defer logger.Log.Sync()

	// Конфиг
	cfg, err := config.ParseFlags(ctx)
	if err != nil {
		return fmt.Errorf("Run: parse flags failed %w", err)
	}

	// База данных
	db, err := db.Init(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("Run: database initialization failed %w", err)
	}

	// Сервисы
	userService := user.NewUserService(userRepo.New(db))

	// Контроллеры
	user := userHandler.NewHandler(cfg, userService)
	server := handlers.NewController(user)

	// Роутер
	r := chi.NewRouter()
	r.Use(middlewares.Recovery)
	r.Mount("/", server.BuildRoute(ctx))

	logger.Log.Info("Running server", zap.String("address", cfg.Address))

	return http.ListenAndServe(cfg.Address, r)
}
