package order

import (
	"context"
	"time"
)

type Order struct {
	ID       int       `json:"id,omitempty"`
	Number   string    `json:"number,omitempty"`
	UserID   int       `json:"user_id,omitempty"`
	Status   string    `json:"status,omitempty"`
	Accrual  int       `json:"accrual,omitempty"`
	Uploaded time.Time `json:"uploaded_at,omitempty"`
}

type Service interface {
	Upload(ctx context.Context, order *Order) error
	List(ctx context.Context, userID int) ([]*Order, error)
}

type Repository interface {
	SaveOrder(ctx context.Context, order *Order) error
	GetOrders(ctx context.Context, userID int) ([]*Order, error)
}
