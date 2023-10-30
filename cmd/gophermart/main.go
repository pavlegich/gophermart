package main

import (
	"net/http"

	"github.com/pavlegich/gophermart/internal/app"
	"github.com/pavlegich/gophermart/internal/infra/logger"
	"go.uber.org/zap"
)

func main() {
	done := make(chan bool, 1)
	if err := app.Run(done); err != http.ErrServerClosed {
		logger.Log.Error("main: run app failed",
			zap.Error(err))
	}
	<-done
}
