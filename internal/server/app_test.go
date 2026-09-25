package server_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/go-logr/zapr"

	"github.com/nikimonax/laretz/internal/lib/browser"
	"github.com/nikimonax/laretz/internal/server"
	"github.com/nikimonax/laretz/internal/testing/httpx"
)

type fakeOpener struct {
	opened bool
	url    string
}

func (f *fakeOpener) Open(url string) error {
	f.opened = true
	f.url = url
	return nil
}

func TestAppStartsServesAndOpensBrowser(t *testing.T) {
	core, records := observer.New(zapcore.DebugLevel)
	logger := zapr.NewLogger(zap.New(core))
	opener := &fakeOpener{}
	laretz := server.NewApp(server.Options{Port: 0, Logger: logger, Opener: opener})
	ctx := context.Background()

	require.NoError(t, laretz.Start(ctx))

	require.True(t, opener.opened, "browser opener must be invoked on start")
	require.True(t, strings.HasPrefix(opener.url, "http://127.0.0.1:"), "opener must receive the loopback URL")

	get := func(url string) (*http.Response, error) {
		client := &http.Client{Timeout: 2 * time.Second}
		return client.Do(httpx.NewRequest(t, http.MethodGet, url))
	}

	res, err := get(opener.url)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()
	require.Equal(t, http.StatusOK, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Laretz")

	var urlLogged bool
	for _, entry := range records.All() {
		if entry.Message == "laretz is running" {
			for _, f := range entry.Context {
				if f.Key == "url" && f.String == opener.url {
					urlLogged = true
				}
			}
		}
	}
	assert.True(t, urlLogged, "start must log the app URL")

	require.NoError(t, laretz.Stop(ctx))

	_, err = get(opener.url)
	assert.Error(t, err, "server must be shut down after Stop")
}

var _ browser.Opener = (*fakeOpener)(nil)
