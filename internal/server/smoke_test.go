package server_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/server"
	"github.com/nikimonax/laretz/internal/server/config"
	"github.com/nikimonax/laretz/internal/testing/httpx"
)

func freePort(t *testing.T) uint16 {
	t.Helper()

	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	port := uint16(listener.Addr().(*net.TCPAddr).Port)
	require.NoError(t, listener.Close())

	return port
}

func httpClient() *http.Client {
	return &http.Client{Timeout: 5 * time.Second}
}

func fetch(t *testing.T, method string, url string) (int, []byte) {
	t.Helper()

	res, err := httpClient().Do(httpx.NewRequest(t, method, url))
	require.NoError(t, err, "%s %s", method, url)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	return res.StatusCode, body
}

func TestFirstRunSmoke(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.yaml")
	dataDir := filepath.Join(root, "data")

	cfg, err := config.EnsureConfig(configPath)
	require.NoError(t, err)

	cfg.Data.Dir = dataDir
	cfg.Server.Port = freePort(t)

	app, err := server.New(cfg)
	require.NoError(t, err)
	require.NoError(t, app.Err(), "fx app must assemble")

	ctx := context.Background()
	require.NoError(t, app.Start(ctx), "first run must start: migrations and listener")

	defer func() {
		require.NoError(t, app.Stop(ctx))
	}()

	_, err = os.Stat(cfg.Data.DatabaseFile())
	require.NoError(t, err, "migrations must create the sqlite database")

	base := fmt.Sprintf("http://127.0.0.1:%d", cfg.Server.Port)

	tests := []struct {
		path string
		want int
	}{
		{path: "/", want: http.StatusOK},
		{path: "/assets", want: http.StatusOK},
		{path: "/assets/search?q=btc", want: http.StatusOK},
		{path: "/nope", want: http.StatusNotFound},
		{path: "/api/portfolio/999/chart", want: http.StatusNotFound},
	}

	for _, tt := range tests {
		status, body := fetch(t, http.MethodGet, base+tt.path)

		assert.Equal(t, tt.want, status, "GET %s", tt.path)
		assert.NotEmpty(t, body, "GET %s must render a page", tt.path)
	}

	status, _ := fetch(t, http.MethodPatch, base+"/")
	assert.Equal(t, http.StatusMethodNotAllowed, status)

	// second run reads the config written by the first run
	reread, err := config.ReadConfig(configPath)
	require.NoError(t, err, "config written by first run must be readable by second run")
	require.NoError(t, config.ValidateConfig(reread))

	reread.Data.Dir = dataDir
	reread.Server.Port = freePort(t)

	second, err := server.New(reread)
	require.NoError(t, err)
	require.NoError(t, second.Err())

	require.NoError(t, second.Start(ctx))

	status, body := fetch(t, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d", reread.Server.Port)+"/")

	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, body)

	require.NoError(t, second.Stop(ctx))
}
