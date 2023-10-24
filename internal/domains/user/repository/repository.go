package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pavlegich/gophermart/internal/domains/user"

	errs "github.com/pavlegich/gophermart/internal/errors"
)

type Repository struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// GetUserByID возвращает конкретного пользователя из хранилища
func (r *Repository) GetUserByLogin(ctx context.Context, login string) (*user.User, error) {
	// Проверка базы данных
	if err := r.db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("GetUserByLogin: connection to database is died %w", err)
	}

	// Выполнение запроса на получение строки с данными пользователя
	row := r.db.QueryRowContext(ctx, "SELECT id, login, password FROM users WHERE login = $1", login)

	// Запись данных пользователя в структуру
	var user user.User
	err := row.Scan(&user.ID, &user.Login, &user.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("GetUserByLogin: scan row failed %w", errs.ErrUserNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("GetUserByLogin: scan row failed %w", err)
	}

	err = row.Err()
	if err != nil {
		return nil, fmt.Errorf("GetUserByLogin: row.Err %w", err)
	}

	return &user, nil
}

// Save сохраняет данные пользователя в хранилище
func (r *Repository) SaveUser(ctx context.Context, u *user.User) error {
	// Проверка базы данных
	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("SaveUser: connection to database in died %w", err)
	}

	// Начало транзакции
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("SaveUser: begin transaction failed %w", err)
	}
	defer tx.Rollback()

	// ======================
	// При запросе к БД, если пользователь создан, то все равно будет ошибка
	// Вот ее и нужно обработать
	// ======================

	// Проверка отсутствия пользователя
	id := tx.QueryRowContext(ctx, "SELECT id FROM users WHERE login = $1", u.Login)
	var storedID int
	err = id.Scan(&storedID)
	if err == nil {
		return fmt.Errorf("SaveUser: scan row with user id failed %w", errs.ErrLoginBusy)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("SaveUser: scan row with user id failed %w", err)
	}

	// Выполнение запроса к базе данных
	if _, err := tx.ExecContext(ctx, "INSERT INTO users (login, password) VALUES ($1, $2)",
		u.Login, u.Password); err != nil {
		return fmt.Errorf("SaveUser: insert into table failed %w", err)
	}

	// Проверка присутствия пользователя
	id = tx.QueryRowContext(ctx, "SELECT id FROM users WHERE login = $1", u.Login)
	if err := id.Scan(&storedID); err != nil {
		return fmt.Errorf("SaveUser: saved user not found in table %w", err)
	}
	u.ID = storedID

	// Подтверждение транзакции
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("SaveUser: commit transaction failed %w", err)
	}

	return nil
}
