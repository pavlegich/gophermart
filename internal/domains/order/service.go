package order

import (
	"context"

	errs "github.com/pavlegich/gophermart/internal/errors"
	"github.com/pavlegich/gophermart/internal/utils"
)

// OrderService содержит интерфефсы и данные сервиса заказов
type OrderService struct {
	repo Repository
}

// NewOrderService возвращает новый сервис для заказов
func NewOrderService(repo Repository) *OrderService {
	return &OrderService{
		repo: repo,
	}
}

// Upload обрабатывает и сохраняет заказ в хранилище
func (s *OrderService) Upload(ctx context.Context, order *Order) error {
	if !utils.LuhnValid(order.Number) {
		return errs.ErrIncorrectNumberFormat
	}
	if err := s.repo.SaveOrder(ctx, order); err != nil {
		return err
	}
	return nil
}
