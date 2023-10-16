package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pavlegich/gophermart/internal/domains/balance"
)

type Repository struct {
	db *sql.DB
}

func NewBalanceRepo(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// GetBalanceActions возвращает список операций для баланса пользователя
func (r *Repository) GetBalanceActions(ctx context.Context, userID int) ([]*balance.Balance, error) {
	// Проверка базы данных
	if err := r.db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("GetBalanceActions: connection to database in died %w", err)
	}

	// Начало транзакции
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("GetBalanceActions: begin transaction failed %w", err)
	}
	defer tx.Rollback()

	// Получение данных заказа
	rows, err := tx.QueryContext(ctx, "SELECT id, action, amount, user_id, order_id, created_at "+
		"FROM balances WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("GetBalanceActions: read rows from table failed %w", err)
	}
	defer rows.Close()

	storedBalance := make([]*balance.Balance, 0)
	for rows.Next() {
		var bal balance.Balance
		err = rows.Scan(&bal.ID, &bal.Action, &bal.Amount, &bal.UserID, &bal.OrderID, &bal.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("GetBalanceActions: scan row failed %w", err)
		}
		storedBalance = append(storedBalance, &bal)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("GetBalanceActions: rows.Err %w", err)
	}

	// Подтверждение транзакции
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("GetBalanceActions: commit transaction failed %w", err)
	}

	return storedBalance, nil
}
