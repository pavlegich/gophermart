package errors

import (
	"errors"
)

var (
	ErrLoginBusy        = errors.New("login is busy")
	ErrUserNotFound     = errors.New("user not found")
	ErrPasswordNotMatch = errors.New("passwords do not match")
	ErrUserUnauthorized = errors.New("user unauthorized")
)
