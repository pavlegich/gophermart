package handlers

import "github.com/pavlegich/gophermart/internal/infra/config"

type Controller struct {
	config *config.Config
}

func NewController(cfg *config.Config) *Controller {
	return &Controller{
		config: cfg,
	}
}

func (c *Controller) GetAddress() string {
	return c.config.GetAddress()
}

func (c *Controller) GetDBuri() string {
	return c.config.GetDBuri()
}

func (c *Controller) GetAccrualAddr() string {
	return c.config.GetAccrualAddr()
}
