package middlewares

import (
	"net/http"

	"github.com/pavlegich/gophermart/internal/infra/logger"
	"go.uber.org/zap"
)

// Recovery восстанавливает работу в случае паники при запуске
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				logger.Log.Error("Recovery: server router panic", zap.Any("error", err))

				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
