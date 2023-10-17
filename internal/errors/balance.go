package errors

import "errors"

var (
	ErrInsufficientFunds   = errors.New("account has insufficient funds")
	ErrOperationsNotFound  = errors.New("balance operations not found")
	ErrWithdrawalsNotFound = errors.New("withdrawals operations not found")
)
