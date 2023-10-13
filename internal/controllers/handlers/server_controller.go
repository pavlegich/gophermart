package handlers

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/gophermart/internal/controllers/middlewares"
	users "github.com/pavlegich/gophermart/internal/domains/user/controllers/http"
	"github.com/pavlegich/gophermart/internal/infra/config"
)

type Controller struct {
	db  *sql.DB
	cfg *config.Config
}

func NewController(db *sql.DB, cfg *config.Config) *Controller {
	return &Controller{
		db:  db,
		cfg: cfg,
	}
}

// BuildRoute регистрирует обработчики и мидлвары в роутере
func (c *Controller) BuildRoute(ctx context.Context) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middlewares.WithLogging)
	users.Activate(r, c.cfg, c.db)
	r.Get("/", c.HandleMain)

	return r
}
