package balance

import "context"

type BalanceService struct {
	repo Repository
}

// NewOrderService возвращает новый сервис для балансов
func NewBalanceService(repo Repository) *BalanceService {
	return &BalanceService{
		repo: repo,
	}
}

// List возвращает список поступлений и снятий для пользователя
func (s *BalanceService) List(ctx context.Context, userID int) ([]*Balance, error) {
	balanceList, err := s.repo.GetBalanceActions(ctx, userID)
	if err != nil {
		return nil, err
	}
	return balanceList, nil
}
