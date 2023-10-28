package utils

import (
	"context"
	"fmt"
)

type contextKey int

const ContextIDKey contextKey = iota

// GetUserIDFromContext возвращает ID пользователя из контекста
func GetUserIDFromContext(ctx context.Context) (int, error) {
	ctxValue := ctx.Value(ContextIDKey)
	if ctxValue == nil {
		return -1, fmt.Errorf("GetUserIDFromContext: get context value failed")
	}
	userID, ok := ctxValue.(int)
	if !ok {
		return -1, fmt.Errorf("GetUserIDFromContext: convert context value into integer failed")
	}
	return userID, nil
}
