package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Init инициализирует базу данных
func Init(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("pgx", path)
	if err != nil {
		return nil, fmt.Errorf("Init: open database failed %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("Init: connection with database died %w", err)
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("Init: goose set dialect failed %w", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return nil, fmt.Errorf("Init: goose up failed %w", err)
	}

	return db, nil
}
