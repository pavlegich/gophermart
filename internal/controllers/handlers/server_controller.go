package handlers

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/gophermart/internal/controllers/middlewares"
	user "github.com/pavlegich/gophermart/internal/domains/user/controllers/http"
)

type Controller struct {
	user *user.Handler
}

func NewController(user *user.Handler) *Controller {
	return &Controller{
		user: user,
	}
}

// Route регистрирует обработчики и мидлвары в роутере
func (c *Controller) BuildRoute(ctx context.Context) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middlewares.WithLogging)

	r.Get("/", c.HandleMain)
	r.Post("/api/user/register", c.user.HandleRegister)
	r.Post("/api/user/login", c.user.HandleLogin)

	return r
}
