package handlers

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/gophermart/internal/controllers/middlewares"
	balances "github.com/pavlegich/gophermart/internal/domains/balance/controllers/http"
	orders "github.com/pavlegich/gophermart/internal/domains/order/controllers/http"
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
	r.Use(middlewares.WithAuth(c.cfg.JWT))
	r.Use(middlewares.WithCompress)

	r.Get("/", c.HandleMain)

	users.Activate(r, c.cfg, c.db)
	orders.Activate(ctx, r, c.cfg, c.db)
	balances.Activate(r, c.cfg, c.db)

	return r
}
