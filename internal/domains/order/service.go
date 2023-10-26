package order

import (
	"context"
	"fmt"
	"strconv"

	errs "github.com/pavlegich/gophermart/internal/errors"
	"github.com/pavlegich/gophermart/internal/utils"
)

type OrderService struct {
	repo Repository
}

// NewOrderService возвращает новый сервис для заказов
func NewOrderService(repo Repository) *OrderService {
	return &OrderService{
		repo: repo,
	}
}

// Create обрабатывает и сохраняет новый заказ в хранилище
func (s *OrderService) Create(ctx context.Context, ord *Order) error {
	// Проверка корректности номера заказа
	orderNumber, err := strconv.Atoi(ord.Number)
	if err != nil {
		return fmt.Errorf("Create: convert into integer failed %w", errs.ErrIncorrectNumberFormat)
	}
	if !utils.LuhnValid(orderNumber) {
		return fmt.Errorf("Create: luhn check failed %w", errs.ErrIncorrectNumberFormat)
	}
	// Создание нового заказа
	if err := s.repo.CreateOrder(ctx, ord); err != nil {
		return fmt.Errorf("Create: create order failed %w", err)
	}
	return nil
}

// List возвращает список заказов для пользователя
func (s *OrderService) List(ctx context.Context, userID int) ([]*Order, error) {
	orders, err := s.repo.GetAllOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("List: get orders list failed %w", err)
	}
	if len(orders) == 0 {
		return nil, fmt.Errorf("List: no orders found failed %w", errs.ErrOrdersNotFound)
	}
	return orders, nil
}

// Upload обрабатывает и сохраняет заказ в хранилище
func (s *OrderService) Upload(ctx context.Context, ord *Order) error {
	orderNumber, err := strconv.Atoi(ord.Number)
	if err != nil {
		return fmt.Errorf("Upload: convert into integer failed %w", errs.ErrIncorrectNumberFormat)
	}
	if !utils.LuhnValid(orderNumber) {
		return fmt.Errorf("Upload: luhn check failed %w", errs.ErrIncorrectNumberFormat)
	}
	if err := s.repo.UpdateOrder(ctx, ord); err != nil {
		return fmt.Errorf("Upload: save order failed %w", err)
	}
	return nil
}

// ListUnprocessed возвращает список всех ещё необработанных заказов
func (s *OrderService) ListUnprocessed(ctx context.Context) ([]*Order, error) {
	orders, err := s.repo.GetUnprocessedOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListUnprocessed: get orders list failed %w", err)
	}
	if len(orders) == 0 {
		return nil, fmt.Errorf("ListUnprocessed: no orders found failed %w", errs.ErrOrdersNotFound)
	}
	return orders, nil
}
