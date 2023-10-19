package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pavlegich/gophermart/internal/domains/balance"
	errs "github.com/pavlegich/gophermart/internal/errors"
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
func (r *Repository) GetBalanceOperations(ctx context.Context, userID int) ([]*balance.Balance, error) {
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
	rows, err := tx.QueryContext(ctx, "SELECT id, action, amount, user_id, order_number, created_at "+
		"FROM balances WHERE user_id = $1 ORDER BY created_at DESC;", userID)
	if err != nil {
		return nil, fmt.Errorf("GetBalanceActions: read rows from table failed %w", err)
	}
	defer rows.Close()

	storedBalance := make([]*balance.Balance, 0)
	for rows.Next() {
		var bal balance.Balance
		err = rows.Scan(&bal.ID, &bal.Action, &bal.Amount, &bal.UserID, &bal.Order, &bal.CreatedAt)
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

func (r *Repository) UploadWithdrawal(ctx context.Context, bal *balance.Balance) error {
	// Проверка базы данных
	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("UploadWithdrawal: connection to database in died %w", err)
	}

	// Начало транзакции
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("UploadWithdrawal: begin transaction failed %w", err)
	}
	defer tx.Rollback()

	// Расчёт текущего баланса
	rows, err := tx.QueryContext(ctx, "SELECT action, amount FROM balances "+
		"WHERE user_id = $1 FOR UPDATE", bal.UserID)
	if err != nil {
		return fmt.Errorf("UploadWithdrawal: user opertations get failed %w", err)
	}
	var uBalance float32 = 0
	for rows.Next() {
		var uOp struct {
			action string
			amount float32
		}
		if err := rows.Scan(&uOp.action, &uOp.amount); err != nil {
			return fmt.Errorf("UploadWithdrawal: scan operation rows failed %w", err)
		}

		switch uOp.action {
		case "ACCRUAL":
			uBalance += uOp.amount
		case "WITHDRAWAL":
			uBalance -= uOp.amount
		}
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("UploadWithdrawal: rows.Err %w", err)
	}

	if uBalance-bal.Amount < 0 {
		return fmt.Errorf("UploadWithdrawal: %w", errs.ErrInsufficientFunds)
	}

	// Подготовка запроса для вставки строки с операцией
	statement, err := tx.PrepareContext(ctx, "INSERT INTO balances "+
		"(action, amount, user_id, order_number) VALUES ($1, $2, $3, $4);")
	if err != nil {
		return fmt.Errorf("UploadWithdrawal: prepare statement failed %w", err)
	}
	defer statement.Close()

	// Исполнение запроса к базе данных
	if _, err := statement.ExecContext(ctx, bal.Action, bal.Amount,
		bal.UserID, bal.Order); err != nil {
		return fmt.Errorf("UploadWithdrawal: statement exec failed %w", err)
	}

	// Подтверждение транзакции
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("UploadWithdrawal: commit transaction failed %w", err)
	}

	return nil
}
