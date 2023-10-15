package errors

import (
	"errors"
)

var (
	ErrOrderAlreadyUpload    = errors.New("order already uploaded by this user")
	ErrOrderUploadByAnother  = errors.New("order already uploaded by another user")
	ErrIncorrectNumberFormat = errors.New("order has incorrect number format")
)
