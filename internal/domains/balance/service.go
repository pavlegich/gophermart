package balance

import (
	"context"
	"fmt"
	"strconv"

	errs "github.com/pavlegich/gophermart/internal/errors"
	"github.com/pavlegich/gophermart/internal/utils"
)

type BalanceService struct {
	repo Repository
}

func NewBalanceService(repo Repository) *BalanceService {
	return &BalanceService{
		repo: repo,
	}
}

// List возвращает список поступлений и снятий для пользователя
func (s *BalanceService) List(ctx context.Context, userID int) ([]*Balance, error) {
	balanceList, err := s.repo.GetBalanceOperations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("List: get balance operations failed %w", err)
	}
	if len(balanceList) == 0 {
		return nil, fmt.Errorf("List: %w", errs.ErrOperationsNotFound)
	}
	return balanceList, nil
}

// Withdraw обрабатывает списание баллов
func (s *BalanceService) Withdraw(ctx context.Context, b *Balance) error {
	orderNumber, err := strconv.Atoi(b.Order)
	if err != nil {
		return fmt.Errorf("Withdraw: convert into integer failed %w", errs.ErrIncorrectNumberFormat)
	}
	if !utils.LuhnValid(orderNumber) {
		return fmt.Errorf("Withdraw: luhn check failed %w", errs.ErrIncorrectNumberFormat)
	}
	if err := s.repo.UploadWithdrawal(ctx, b); err != nil {
		return fmt.Errorf("Withdraw: upload withdrawal failed %w", err)
	}
	return nil
}
