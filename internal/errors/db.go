package errors

import (
	"errors"
)

var (
	ErrLoginBusy = errors.New("login is busy")
)
