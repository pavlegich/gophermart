package logger

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type (
	// ResponseData хранит сведения об ответе
	ResponseData struct {
		Status int
		Size   int
		Body   *bytes.Buffer
	}

	// LoggingResponseWriter реализует http.ResponseWriter
	LoggingResponseWriter struct {
		http.ResponseWriter
		ResponseData *ResponseData
	}
)

var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(ctx context.Context, level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return fmt.Errorf("Initialize: parse level failed %w", err)
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("Initialize: logger build failed %w", err)
	}
	Log = zl
	return nil
}

// Переопределение метода WriteHeader
func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.Status = statusCode
}

// Переопределение метода Write
func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return size, fmt.Errorf("Write: response write %w", err)
	}
	r.ResponseData.Size += size
	r.ResponseData.Body.Write(b)
	return size, nil
}
