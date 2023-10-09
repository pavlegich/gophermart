package middlewares

import (
	"bytes"
	"net/http"
	"time"

	"github.com/pavlegich/gophermart/internal/infra/logger"
	"go.uber.org/zap"
)

func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &logger.ResponseData{
			Status: 0,
			Size:   0,
			Body:   bytes.NewBufferString(""),
		}
		lw := logger.LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		logger.Log.Info("incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", duration),
			zap.Int("status", responseData.Status),
			zap.Int("size", responseData.Size),
			zap.String("body", responseData.Body.String()),
		)
	}
	return http.HandlerFunc(logFn)
}
