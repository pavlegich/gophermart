package middlewares

import (
	"net/http"

	"github.com/pavlegich/gophermart/internal/infra/logger"
	"go.uber.org/zap"
)

// Recovery восстанавливает работу в случае паники при запуске
func Recovery(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				logger.Log.Info("Recovery: server router panic", zap.Any("error", err))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}
