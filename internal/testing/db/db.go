// Package db provides migrated SQLite test databases.
package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/impl/database"
)

// New opens a migrated SQLite database in a fresh temporary directory. The
// connection is closed automatically when the test finishes.
func New(t testing.TB) *sql.DB {
	t.Helper()

	conn, err := database.Connect(filepath.Join(t.TempDir(), "laretz.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	require.NoError(t, database.Migrate(context.Background(), conn))

	return conn
}
