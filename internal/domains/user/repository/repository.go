package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pavlegich/gophermart/internal/domains/user"
	"golang.org/x/crypto/bcrypt"

	errs "github.com/pavlegich/gophermart/internal/errors"
)

// Reposity содержит указатель на базу данных
type Repository struct {
	db *sql.DB
}

// New создает новый repository для пользователя
func New(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetUserByID(ctx context.Context, login string) (*user.User, error) {
	if err := r.db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("GetUserByID: connection to database is died %w", err)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("GetUserByID: begin transaction failed %w", err)
	}
	defer tx.Rollback()

	row, err := tx.QueryContext(ctx, "SELECT login, password FROM users WHERE login = $1", login)
	if err != nil {
		return nil, fmt.Errorf("GetUserByID: read row from table failed %w", err)
	}
	defer row.Close()

	var user user.User
	row.Next()
	if err := row.Scan(&user.Login, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrUserNotFound
		} else {
			return nil, fmt.Errorf("GetUserByID: scan row failed %w", err)
		}
	}

	err = row.Err()
	if err != nil {
		return nil, fmt.Errorf("GetUserByID: row.Err %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("GetUserByID: commit transaction failed %w", err)
	}

	return &user, nil
}

// GetUsers возвращает список пользователей
func (r *Repository) GetUsers(ctx context.Context) ([]*user.User, error) {
	return nil, nil
}

// Save сохраняет данные пользователя в хранилище
func (r *Repository) SaveUser(ctx context.Context, user *user.User) error {
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

	// Проверка отсутствия пользователя
	id := tx.QueryRowContext(ctx, "SELECT id FROM users WHERE login = $1", user.Login)
	var tmp int
	if err := id.Scan(&tmp); err != sql.ErrNoRows {
		if err == nil {
			return errs.ErrLoginBusy
		} else {
			return fmt.Errorf("SaveUser: query row failed %w", err)
		}
	}

	// Подготовка запроса к базе данных
	statement, err := tx.PrepareContext(ctx, "INSERT INTO users (login, password) VALUES ($1, $2)")
	if err != nil {
		return fmt.Errorf("SaveUser: insert into table failed %w", err)
	}
	defer statement.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("SaveUser: hash generate failed %w", err)
	}

	// Исполнение запроса к базе данных
	if _, err := statement.ExecContext(ctx, user.Login, hashedPassword); err != nil {
		return fmt.Errorf("SaveUser: statement exec failed %w", err)
	}

	// Подтверждение транзакции
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("SaveUser: commit transaction failed %w", err)
	}

	return nil
}
