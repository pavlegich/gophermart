package repository

import (
	"context"
	"database/sql"
	"errors"
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

	// ======================
	// Создать индекс, тогда поиск будет происходить быстрее
	// Почитать про внутрянку работы индексов, спрашивают на собесах
	// ======================

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

	// ======================
	// Обработать ошибку при запросе к БД,
	// он увидит, что строчка уже есть
	// ======================

	// Проверка отсутствия заказа
	userID := tx.QueryRowContext(ctx, "SELECT user_id FROM orders WHERE number = $1", ord.Number)
	var storedUserID int
	err = userID.Scan(&storedUserID)
	if err == nil {
		if ord.UserID == storedUserID {
			return fmt.Errorf("CrateOrder: %w", errs.ErrOrderAlreadyUpload)
		}
		return fmt.Errorf("CrateOrder: %w", errs.ErrOrderUploadByAnother)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("CreateOrder: query row failed %w", err)
	}

	// Выполнение запроса к базе данных
	if _, err := tx.ExecContext(ctx, "INSERT INTO orders (number, user_id) VALUES ($1, $2)",
		ord.Number, ord.UserID); err != nil {
		return fmt.Errorf("CreateOrder: insert into table failed %w", err)
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

	// ======================
	// Сделать эту проверку в самом запросе к БД,
	// и тогда они не будут обрабатываться
	// ======================

	// Проверка отсутствия обработки заказа
	ordStatus := tx.QueryRowContext(ctx, "SELECT status FROM orders WHERE id = $1", ord.ID)
	var storedStatus string
	if err := ordStatus.Scan(&storedStatus); err != nil {
		return fmt.Errorf("SaveOrder: scan row with status failed %w", err)
	}

	if storedStatus == "INVALID" || storedStatus == "PROCESSED" {
		return fmt.Errorf("SaveOrder: order check failed %w", errs.ErrOrderAlreadyProcessed)
	}

	// Выполнение запроса к базе данных
	if _, err := tx.ExecContext(ctx, "UPDATE orders SET status = $1, "+
		"accrual = $2 WHERE id = $3", ord.Status, ord.Accrual, ord.ID); err != nil {
		return fmt.Errorf("SaveOrder: update table failed %w", err)
	}

	// ======================
	// Сделать это отдельным методом,
	// вызывать методы из сервиса с объявлением транзакции там
	// ======================

	// Сохранение информации о начислении, если заказ обработан
	if ord.Status == "PROCESSED" {
		// Проверка отсутствия заказа
		userID := tx.QueryRowContext(ctx, `SELECT user_id FROM balances WHERE order_number = $1 
		AND action = 'ACCRUAL'`, ord.Number)
		var storedUserID int
		err := userID.Scan(&storedUserID)
		if err == nil {
			if ord.UserID == storedUserID {
				return fmt.Errorf("SaveOrder: %w", errs.ErrOrderAlreadyUpload)
			}
			return fmt.Errorf("SaveOrder: %w", errs.ErrOrderUploadByAnother)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("SaveOrder: get order from table balances failed %w", err)
		}

		// Выполнение запроса к базе данных
		if _, err := tx.ExecContext(ctx, `INSERT INTO balances 
		(action, amount, user_id, order_number) VALUES ('ACCRUAL', $1, $2, $3)`,
			ord.Accrual, ord.UserID, ord.Number); err != nil {
			return fmt.Errorf("SaveOrder: insert into balances failed %w", err)
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

	// Получение данных заказа
	rows, err := r.db.QueryContext(ctx, `SELECT id, number, user_id, status, accrual, created_at FROM orders 
	WHERE status = 'NEW' OR status = 'PROCESSING' LIMIT 20`)
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

	return storedOrders, nil
}
