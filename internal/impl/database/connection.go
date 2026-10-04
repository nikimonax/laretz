package database

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"modernc.org/sqlite"
)

var (
	connectParams = []string{
		"busy_timeout(5000)",
		"journal_mode(WAL)",
		"synchronous(NORMAL)",
		"foreign_keys(1)",
	}

	runtimeParams = []string{
		"PRAGMA cache_size = -4000;",
		"PRAGMA temp_store = MEMORY;",
	}
)

func Connect(dbFile string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbFile), 0755); err != nil {
		return nil, err
	}

	dsn := url.URL{
		Scheme:   "file",
		Path:     dbFile,
		RawQuery: url.Values{"_pragma": connectParams}.Encode(),
		OmitHost: true,
	}

	connector, err := sqlite.NewConnector(dsn.String())

	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(0)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		err = errors.Join(err, db.Close())
		return nil, err
	}

	for _, pragma := range runtimeParams {
		if _, err := db.ExecContext(ctx, pragma); err != nil {
			err = errors.Join(err, db.Close())
			return nil, err
		}
	}

	return db, nil
}
