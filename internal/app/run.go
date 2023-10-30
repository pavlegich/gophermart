package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "go.uber.org/automaxprocs"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/gophermart/internal/controllers/handlers"
	"github.com/pavlegich/gophermart/internal/controllers/middlewares"
	"github.com/pavlegich/gophermart/internal/infra/config"
	"github.com/pavlegich/gophermart/internal/infra/database"
	"github.com/pavlegich/gophermart/internal/infra/logger"
	"go.uber.org/zap"
)

// Run инициализирует основные компоненты и запускает сервер
func Run(done chan bool) error {
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

	// База данных
	db, err := database.Init(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("Run: database initialization failed %w", err)
	}
	defer db.Close()

	// Контроллер
	server := handlers.NewController(db, cfg)
	serverRouter := server.BuildRoute(ctx)

	// Роутер
	r := chi.NewRouter()
	r.Use(middlewares.Recovery)
	r.Mount("/", serverRouter)

	// Сервер
	srv := http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	logger.Log.Info("running server", zap.String("addr", cfg.Address))

	// Завершение программы
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		if err := srv.Shutdown(ctx); err != nil {
			logger.Log.Error("server shutdown failed",
				zap.Error(err))
		}
		logger.Log.Info("shutting down gracefully",
			zap.String("signal", sig.String()))
		done <- true
	}()

	return srv.ListenAndServe()
}
