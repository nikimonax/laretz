package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrepareAppFirstRun(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("LARETZ_CONFIG", configPath)

	app, err := prepareApp(Build())

	require.NoError(t, err, "first run without .env must not fail")
	require.NotNil(t, app)
	require.NoError(t, app.Err(), "fx app must assemble")

	_, err = os.Stat(configPath)
	require.NoError(t, err, "config must be created at LARETZ_CONFIG path")
}
