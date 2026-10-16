package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(ctx context.Context, db *sql.DB) error {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	goose.SetBaseFS(migrations)
	defer goose.SetBaseFS(nil)

	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("apply sqlite migrations: %w", err)
	}

	return nil
}
