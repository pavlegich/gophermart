package utils

import (
	"errors"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"golang.org/x/net/context"
)

// GetRequestWithRetry делает повторные запросы при необходимости и возвращает ответ
func GetRequestWithRetry(ctx context.Context, r *http.Request) (*http.Response, error) {
	var err error = nil
	var resp *http.Response
	intervals := []time.Duration{0, time.Second, 3 * time.Second, 5 * time.Second}
	for _, interval := range intervals {
		time.Sleep(interval)
		resp, err = http.DefaultClient.Do(r)
		if !errors.Is(err, syscall.ECONNREFUSED) {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("GetRequestWithRetry: request failed %w", err)
	}
	return resp, nil
}
