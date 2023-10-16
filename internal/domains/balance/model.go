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
	OrderID   int       `json:"order_id"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type Service interface {
	List(ctx context.Context, userID int) ([]*Balance, error)
}

type Repository interface {
	GetBalanceActions(ctx context.Context, userID int) ([]*Balance, error)
}
