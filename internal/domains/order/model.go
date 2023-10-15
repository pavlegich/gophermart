package order

import (
	"context"
	"time"
)

type Order struct {
	ID       int       `json:"id"`
	Number   int       `json:"number,omitempty"`
	UserID   int       `json:"user_id"`
	Status   string    `json:"status"`
	Uploaded time.Time `json:"time"`
}

type Service interface {
	Upload(ctx context.Context, order *Order) error
}

type Repository interface {
	SaveOrder(ctx context.Context, order *Order) error
}
