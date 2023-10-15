package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pavlegich/gophermart/internal/domains/order"
	errs "github.com/pavlegich/gophermart/internal/errors"
)

// Reposity содержит указатель на базу данных
type Repository struct {
	db *sql.DB
}

// New создает новый repository для пользователя
func NewOrderRepo(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// SaveOrder сохраняет данные заказа в хранилище
func (r *Repository) SaveOrder(ctx context.Context, ord *order.Order) error {
	// Проверка базы данных
	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("SaveOrder: connection to database in died %w", err)
	}

	// Начало транзакции
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("SaveOrder: begin transaction failed %w", err)
	}
	defer tx.Rollback()

	// Проверка отсутствия заказа
	userID := tx.QueryRowContext(ctx, "SELECT user_id FROM orders WHERE number = $1", ord.Number)
	var tmp int
	if err := userID.Scan(&tmp); err != sql.ErrNoRows {
		if err == nil {
			if ord.UserID == tmp {
				return errs.ErrOrderAlreadyUpload
			} else {
				return errs.ErrOrderUploadByAnother
			}
		} else {
			return fmt.Errorf("SaveOrder: query row failed %w", err)
		}
	}

	// Подготовка запроса к базе данных
	statement, err := tx.PrepareContext(ctx, "INSERT INTO orders (number, user_id, status) VALUES ($1, $2, $3)")
	if err != nil {
		return fmt.Errorf("SaveOrders: insert into table failed %w", err)
	}
	defer statement.Close()

	// Исполнение запроса к базе данных
	if _, err := statement.ExecContext(ctx, ord.Number, ord.UserID, "NEW"); err != nil {
		return fmt.Errorf("SaveOrder: statement exec failed %w", err)
	}

	// Проверка присутствия заказа
	row := tx.QueryRowContext(ctx, "SELECT id, number, user_id, status, created_at FROM orders WHERE number = $1", ord.Number)
	var tmpOrder order.Order
	if err := row.Scan(&tmpOrder.ID, &tmpOrder.Number, &tmpOrder.UserID, &tmpOrder.Status, &tmpOrder.Uploaded); err != nil {
		return fmt.Errorf("SaveOrder: save order not found in table %w", err)
	}
	ord.ID = tmpOrder.ID
	ord.UserID = tmpOrder.UserID
	ord.Status = tmpOrder.Status
	ord.Uploaded = tmpOrder.Uploaded

	// Подтверждение транзакции
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("SaveOrder: commit transaction failed %w", err)
	}

	return nil
}
