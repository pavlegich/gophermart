package handlers

import (
	"database/sql"

	"github.com/pavlegich/gophermart/internal/infra/config"
)

type Controller struct {
	Config   *config.Config
	Database *sql.DB
}

func NewController(cfg *config.Config, db *sql.DB) *Controller {
	return &Controller{
		Config:   cfg,
		Database: db,
	}
}
