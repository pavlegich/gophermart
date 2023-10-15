package order

import (
	"context"
	"strconv"

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
	orderNumber, err := strconv.Atoi(order.Number)
	if err != nil {
		return errs.ErrIncorrectNumberFormat
	}
	if !utils.LuhnValid(orderNumber) {
		return errs.ErrIncorrectNumberFormat
	}
	if err := s.repo.SaveOrder(ctx, order); err != nil {
		return err
	}
	return nil
}

// List возвращает список заказов для пользователя
func (s *OrderService) List(ctx context.Context, userID int) ([]*Order, error) {
	orders, err := s.repo.GetOrders(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, errs.ErrOrdersNotFound
	}
	return orders, nil
}
