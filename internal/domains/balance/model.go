package balance

import (
	"context"
	"time"
)

type Balance struct {
	ID        int       `json:"id,omitempty"`
	Action    string    `json:"action"`
	Amount    float32   `json:"amount"`
	UserID    int       `json:"user_id,omitempty"`
	Order     string    `json:"order"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type Service interface {
	List(ctx context.Context, userID int) ([]*Balance, error)
	Withdraw(ctx context.Context, balance *Balance) error
}

type Repository interface {
	GetBalanceOperations(ctx context.Context, userID int) ([]*Balance, error)
	UploadWithdrawal(ctx context.Context, balance *Balance) error
}
