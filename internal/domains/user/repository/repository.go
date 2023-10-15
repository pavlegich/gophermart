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

	// Начало транзакции
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("GetUserByLogin: begin transaction failed %w", err)
	}
	defer tx.Rollback()

	// Выполнение запроса на получение строки с данными пользователя
	row := tx.QueryRowContext(ctx, "SELECT id, login, password FROM users WHERE login = $1", login)

	// Запись данных пользователя в структуру
	var user user.User
	if err := row.Scan(&user.ID, &user.Login, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrUserNotFound
		} else {
			return nil, fmt.Errorf("GetUserByLogin: scan row failed %w", err)
		}
	}

	err = row.Err()
	if err != nil {
		return nil, fmt.Errorf("GetUserByLogin: row.Err %w", err)
	}

	// Подтверждение транзакции
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("GetUserByLogin: commit transaction failed %w", err)
	}

	return &user, nil
}

// GetUsers возвращает список пользователей
func (r *Repository) GetUsers(ctx context.Context) ([]*user.User, error) {
	return nil, nil
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

	// Проверка отсутствия пользователя
	id := tx.QueryRowContext(ctx, "SELECT id FROM users WHERE login = $1", u.Login)
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("SaveUser: hash generate failed %w", err)
	}

	// Исполнение запроса к базе данных
	if _, err := statement.ExecContext(ctx, u.Login, hashedPassword); err != nil {
		return fmt.Errorf("SaveUser: statement exec failed %w", err)
	}

	// Проверка присутствия пользователя
	id = tx.QueryRowContext(ctx, "SELECT id FROM users WHERE login = $1", u.Login)
	if err := id.Scan(&tmp); err != nil {
		return fmt.Errorf("SaveUser: saved user not found in table %w", err)
	}
	u.ID = tmp

	// Подтверждение транзакции
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("SaveUser: commit transaction failed %w", err)
	}

	return nil
}
