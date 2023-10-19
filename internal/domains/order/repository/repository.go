package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pavlegich/gophermart/internal/domains/order"
	errs "github.com/pavlegich/gophermart/internal/errors"
)

type Repository struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// GetAllOrders возвращает список заказов для пользователя
func (r *Repository) GetAllOrders(ctx context.Context, userID int) ([]*order.Order, error) {
	// Проверка базы данных
	if err := r.db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("GetAllOrders: connection to database in died %w", err)
	}

	// Начало транзакции
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("GetAllOrders: begin transaction failed %w", err)
	}
	defer tx.Rollback()

	// Получение данных заказа
	rows, err := tx.QueryContext(ctx, "SELECT id, number, user_id, status, accrual, created_at FROM orders WHERE user_id = $1 "+
		"ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, fmt.Errorf("GetAllOrders: read rows from table failed %w", err)
	}
	defer rows.Close()

	storedOrders := make([]*order.Order, 0)
	for rows.Next() {
		var ord order.Order
		if err := rows.Scan(&ord.ID, &ord.Number, &ord.UserID, &ord.Status, &ord.Accrual, &ord.CreatedAt); err != nil {
			return nil, fmt.Errorf("GetAllOrders: scan row failed %w", err)
		}
		storedOrders = append(storedOrders, &ord)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("GetAllOrders: rows.Err %w", err)
	}

	// Подтверждение транзакции
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("GetAllOrders: commit transaction failed %w", err)
	}

	return storedOrders, nil
}

// CreateOrder сохраняет данные нового заказа в хранилище
func (r *Repository) CreateOrder(ctx context.Context, ord *order.Order) error {
	// Проверка базы данных
	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("CreateOrder: connection to database in died %w", err)
	}

	// Начало транзакции
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("CreateOrder: begin transaction failed %w", err)
	}
	defer tx.Rollback()

	// Проверка отсутствия заказа
	userID := tx.QueryRowContext(ctx, "SELECT user_id FROM orders WHERE number = $1", ord.Number)
	var storedUserID int
	if err := userID.Scan(&storedUserID); err != sql.ErrNoRows {
		if err == nil {
			if ord.UserID == storedUserID {
				return fmt.Errorf("CrateOrder: %w", errs.ErrOrderAlreadyUpload)
			} else {
				return fmt.Errorf("CrateOrder: %w", errs.ErrOrderUploadByAnother)
			}
		} else {
			return fmt.Errorf("CreateOrder: query row failed %w", err)
		}
	}

	// Подготовка запроса к базе данных
	statement, err := tx.PrepareContext(ctx, "INSERT INTO orders (number, user_id) VALUES ($1, $2)")
	if err != nil {
		return fmt.Errorf("CreateOrder: insert into table failed %w", err)
	}
	defer statement.Close()

	// Исполнение запроса к базе данных
	if _, err := statement.ExecContext(ctx, ord.Number, ord.UserID); err != nil {
		return fmt.Errorf("CreateOrder: statement exec failed %w", err)
	}

	// Проверка присутствия заказа
	row := tx.QueryRowContext(ctx, "SELECT id, number, user_id, status, created_at FROM orders WHERE number = $1", ord.Number)
	var tmpOrder order.Order
	if err := row.Scan(&tmpOrder.ID, &tmpOrder.Number, &tmpOrder.UserID, &tmpOrder.Status, &tmpOrder.CreatedAt); err != nil {
		return fmt.Errorf("CreateOrder: save order not found in table %w", err)
	}
	ord.ID = tmpOrder.ID
	ord.UserID = tmpOrder.UserID
	ord.Status = tmpOrder.Status
	ord.CreatedAt = tmpOrder.CreatedAt

	// Подтверждение транзакции
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("CreateOrder: commit transaction failed %w", err)
	}

	return nil
}

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

	// Проверка отсутствия обработки заказа
	ordStatus := tx.QueryRowContext(ctx, "SELECT status FROM orders WHERE id = $1", ord.ID)
	var storedStatus string
	if err := ordStatus.Scan(&storedStatus); err != nil {
		return fmt.Errorf("SaveOrder: scan row with status failed %w", err)
	}
	if storedStatus == "INVALID" || storedStatus == "PROCESSED" {
		return fmt.Errorf("SaveOrder: order check failed %w", errs.ErrOrderAlreadyProcessed)
	}

	// Подготовка запроса к базе данных
	statement, err := tx.PrepareContext(ctx, "UPDATE orders SET status = $1, "+
		"accrual = $2 WHERE id = $3")
	if err != nil {
		return fmt.Errorf("SaveOrder: update table failed %w", err)
	}
	defer statement.Close()

	// Исполнение запроса к базе данных
	if _, err := statement.ExecContext(ctx, ord.Status, ord.Accrual, ord.ID); err != nil {
		return fmt.Errorf("SaveOrder: statement exec failed %w", err)
	}

	// Сохранение информации о начислении, если заказ обработан
	if ord.Status == "PROCESSED" {
		// Проверка отсутствия заказа
		userID := tx.QueryRowContext(ctx, "SELECT user_id FROM balances WHERE order_number = $1 "+
			"AND action = 'ACCRUAL'", ord.Number)
		var storedUserID int
		if err := userID.Scan(&storedUserID); err != sql.ErrNoRows {
			if err == nil {
				if ord.UserID == storedUserID {
					return fmt.Errorf("SaveOrder: %w", errs.ErrOrderAlreadyUpload)
				} else {
					return fmt.Errorf("SaveOrder: %w", errs.ErrOrderUploadByAnother)
				}
			} else {
				return fmt.Errorf("SaveOrder: get order from table balances failed %w", err)
			}
		}

		// Подготовка запроса к базе данных
		statement, err := tx.PrepareContext(ctx, "INSERT INTO balances "+
			"(action, amount, user_id, order_number) VALUES ('ACCRUAL', $1, $2, $3)")
		if err != nil {
			return fmt.Errorf("SaveOrder: insert into table balances failed %w", err)
		}
		defer statement.Close()

		// Исполнение запроса к базе данных
		if _, err := statement.ExecContext(ctx, ord.Accrual, ord.UserID, ord.Number); err != nil {
			return fmt.Errorf("SaveOrder: statement exec into balances failed %w", err)
		}
	}

	// Подтверждение транзакции
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("SaveOrder: commit transaction failed %w", err)
	}

	return nil
}

// GetUnprocessedOrders возвращает список всех необработанных заказов
func (r *Repository) GetUnprocessedOrders(ctx context.Context) ([]*order.Order, error) {
	// Проверка базы данных
	if err := r.db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("GetUnprocessedOrders: connection to database in died %w", err)
	}

	// Начало транзакции
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("GetUnprocessedOrders: begin transaction failed %w", err)
	}
	defer tx.Rollback()

	// Получение данных заказа
	rows, err := tx.QueryContext(ctx, "SELECT id, number, user_id, status, accrual, created_at FROM orders "+
		"WHERE status = 'NEW' OR status = 'PROCESSING'")
	if err != nil {
		return nil, fmt.Errorf("GetUnprocessedOrders: read rows from table failed %w", err)
	}
	defer rows.Close()

	storedOrders := make([]*order.Order, 0)
	for rows.Next() {
		var ord order.Order
		if err := rows.Scan(&ord.ID, &ord.Number, &ord.UserID, &ord.Status, &ord.Accrual, &ord.CreatedAt); err != nil {
			return nil, fmt.Errorf("GetUnprocessedOrders: scan row failed %w", err)
		}
		storedOrders = append(storedOrders, &ord)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("GetAllOrders: rows.Err %w", err)
	}

	// Подтверждение транзакции
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("GetAllOrders: commit transaction failed %w", err)
	}

	return storedOrders, nil
}
