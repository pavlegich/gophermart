package order

import (
	"context"
	"time"
)

type Order struct {
	ID        int       `json:"id"`
	Number    string    `json:"number"`
	UserID    int       `json:"user_id,omitempty"`
	Status    string    `json:"status,omitempty"`
	Accrual   float32   `json:"accrual,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type Service interface {
	Create(ctx context.Context, order *Order) error
	List(ctx context.Context, userID int) ([]*Order, error)
	Upload(ctx context.Context, order *Order) error
}

type Repository interface {
	CreateOrder(ctx context.Context, order *Order) error
	GetOrders(ctx context.Context, userID int) ([]*Order, error)
	SaveOrder(ctx context.Context, order *Order) error
}
