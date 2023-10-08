package main

import (
	"github.com/pavlegich/gophermart/internal/app"
	"github.com/pavlegich/gophermart/internal/infra/logger"
	"go.uber.org/zap"
)

func main() {
	if err := app.Run(); err != nil {
		logger.Log.Error("main: run app failed",
			zap.Error(err))
	}
}
