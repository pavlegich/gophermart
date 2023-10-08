package app

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/gophermart/internal/controllers/middlewares"
	"github.com/pavlegich/gophermart/internal/infra/logger"
	"go.uber.org/zap"
)

// Run запускает сервер
func Run() error {
	addr := "localhost:8080"

	ctx := context.Background()

	if err := logger.Initialize(ctx, "Info"); err != nil {
		return err
	}
	defer logger.Log.Sync()

	r := chi.NewRouter()
	r.Use(middlewares.Recovery)

	logger.Log.Info("Running server", zap.String("address", addr))

	return http.ListenAndServe(addr, r)
}
