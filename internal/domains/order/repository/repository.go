package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
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

	// Получение данных заказа
	rows, err := r.db.QueryContext(ctx, `SELECT id, number, user_id, status, accrual, created_at 
	FROM orders WHERE user_id = $1 ORDER BY created_at DESC`, userID)
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

	return storedOrders, nil
}

// GetOrderByNumber возвращает данные заказа по его номеру
func (r *Repository) GetOrderByNumber(ctx context.Context, n string) (*order.Order, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, number, user_id, status, created_at FROM orders WHERE number = $1`, n)
	var storedOrder order.Order
	if err := row.Scan(&storedOrder.ID, &storedOrder.Number, &storedOrder.UserID, &storedOrder.Status, &storedOrder.CreatedAt); err != nil {
		return nil, fmt.Errorf("GetOrderByNumber: scan order failed %w", err)
	}
	return &storedOrder, nil
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

	// Выполнение запроса к базе данных и получение данных для заказа
	var storedOrder order.Order
	row := tx.QueryRowContext(ctx, `INSERT INTO orders (number, user_id) VALUES ($1, $2) 
	RETURNING id, number, user_id, status, created_at;`, ord.Number, ord.UserID)
	if err := row.Scan(&storedOrder.ID, &storedOrder.Number, &storedOrder.UserID,
		&storedOrder.Status, &storedOrder.CreatedAt); err != nil {
		storedOrder, err := r.GetOrderByNumber(ctx, ord.Number)
		if err != nil {
			return fmt.Errorf("CreateOrder: get order by number failed %w", err)
		}
		if storedOrder.UserID != ord.UserID {
			return fmt.Errorf("CreateOrder: %w", errs.ErrOrderUploadByAnother)
		}
		if storedOrder.UserID == ord.UserID {
			return fmt.Errorf("CreateOrder: %w", errs.ErrOrderAlreadyUpload)
		}
		return fmt.Errorf("CreateOrder: insert into table failed %w", err)
	}
	ord.ID = storedOrder.ID
	ord.UserID = storedOrder.UserID
	ord.Status = storedOrder.Status
	ord.CreatedAt = storedOrder.CreatedAt

	// Подтверждение транзакции
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("CreateOrder: commit transaction failed %w", err)
	}

	return nil
}

// UpdateOrder обновляет данные о заказе и создаёт запись о начислении за обработанный заказ
func (r *Repository) UpdateOrder(ctx context.Context, ord *order.Order) error {
	// Проверка базы данных
	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("UpdateOrder: connection to database in died %w", err)
	}

	// Начало транзакции
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("UpdateOrder: begin transaction failed %w", err)
	}
	defer tx.Rollback()

	// Выполнение запроса к базе данных
	if _, err := tx.ExecContext(ctx, `UPDATE orders SET status = $1, accrual = $2 
	WHERE id = $3 AND status NOT IN ('PROCESSED', 'INVALID')`,
		ord.Status, ord.Accrual, ord.ID); err != nil {
		return fmt.Errorf("UpdateOrder: update table failed %w", err)
	}

	// Сохранение информации о начислении, если заказ обработан
	if ord.Status == "PROCESSED" {
		if err := r.CreateAccrualForOrder(ctx, ord); err != nil {
			return fmt.Errorf("UpdateOrder: create accrual in balances table for order failed %w", err)
		}
	}

	// Подтверждение транзакции
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("UpdateOrder: commit transaction failed %w", err)
	}

	return nil
}

// CreateAccrualForOrder создаёт начисление по заказу для пользователя в хранилище
func (r *Repository) CreateAccrualForOrder(ctx context.Context, ord *order.Order) error {
	if _, err := r.db.ExecContext(ctx, `INSERT INTO balances 
	(action, amount, user_id, order_number) VALUES ('ACCRUAL', $1, $2, $3)`,
		ord.Accrual, ord.UserID, ord.Number); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("CreateAccrualForOrder: %w", errs.ErrOrderAlreadyProcessed)
		}
		return fmt.Errorf("CreateAccrualForOrder: insert into balances failed %w", err)
	}
	return nil
}

// GetUnprocessedOrders возвращает список всех необработанных заказов
func (r *Repository) GetUnprocessedOrders(ctx context.Context) ([]*order.Order, error) {
	// Проверка базы данных
	if err := r.db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("GetUnprocessedOrders: connection to database in died %w", err)
	}

	// Получение данных заказа
	rows, err := r.db.QueryContext(ctx, `SELECT id, number, user_id, status, accrual, created_at FROM orders 
	WHERE status NOT IN ('PROCESSED', 'INVALID') LIMIT 10`)
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
		return nil, fmt.Errorf("GetUnprocessedOrders: rows.Err %w", err)
	}

	return storedOrders, nil
}
