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
	orderNumber, err := strconv.Atoi(ord.Number)
	if err != nil {
		return fmt.Errorf("Create: convert into integer failed %w", errs.ErrIncorrectNumberFormat)
	}
	if !utils.LuhnValid(orderNumber) {
		return fmt.Errorf("Create: luhn check failed %w", errs.ErrIncorrectNumberFormat)
	}
	if err := s.repo.CreateOrder(ctx, ord); err != nil {
		return fmt.Errorf("Create: create order failed %w", err)
	}
	return nil
}

// List возвращает список заказов для пользователя
func (s *OrderService) List(ctx context.Context, userID int) ([]*Order, error) {
	orders, err := s.repo.GetOrders(ctx, userID)
	if err != nil {
		fmt.Println("ERROR:", err)
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
	if err := s.repo.SaveOrder(ctx, ord); err != nil {
		return fmt.Errorf("Upload: save order failed %w", err)
	}
	return nil
}
