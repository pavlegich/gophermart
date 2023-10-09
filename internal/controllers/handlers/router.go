package handlers

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/gophermart/internal/controllers/middlewares"
)

// Route регистрирует обработчики и мидлвары в роутере
func (c *Controller) Route(ctx context.Context) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middlewares.WithLogging)

	r.Get("/", c.HandleMain)

	return r
}
