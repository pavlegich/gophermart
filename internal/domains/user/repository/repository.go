package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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

// GetUserByLogin возвращает конкретного пользователя из хранилища
func (r *Repository) GetUserByLogin(ctx context.Context, login string) (*user.User, error) {
	// Проверка базы данных
	if err := r.db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("GetUserByLogin: connection to database is died %w", err)
	}

	// Выполнение запроса на получение строки с данными пользователя
	row := r.db.QueryRowContext(ctx, `SELECT id, login, password FROM users WHERE login = $1`, login)

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

// CreateUser сохраняет данные пользователя в хранилище
func (r *Repository) CreateUser(ctx context.Context, u *user.User) error {
	// Проверка базы данных
	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("CreateUser: connection to database in died %w", err)
	}

	// Выполнение запроса к базе данных
	var storedID int
	if err := r.db.QueryRowContext(ctx, `INSERT INTO users (login, password) VALUES ($1, $2) 
	RETURNING id`, u.Login, u.Password).Scan(&storedID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("CreateUser: %w", errs.ErrLoginBusy)
		}
		return fmt.Errorf("CreateUser: insert into table failed %w", err)
	}
	u.ID = storedID

	return nil
}
