package handlers

import (
	"database/sql"

	"github.com/pavlegich/gophermart/internal/domains/user"
	"github.com/pavlegich/gophermart/internal/infra/config"
)

type Controller struct {
	Config   *config.Config
	Database *sql.DB
	User     user.Service
}

func NewController(cfg *config.Config, db *sql.DB, user user.Service) *Controller {
	return &Controller{
		Config:   cfg,
		Database: db,
		User:     user,
	}
}
