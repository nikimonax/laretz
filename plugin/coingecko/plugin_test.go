package coingecko

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/pkg/plugin"
)

type fakeApp struct {
	config string
	logger logr.Logger
	assets plugin.AssetProvider
	quotes plugin.QuoteProvider
}

func (app *fakeApp) Logger() logr.Logger {
	if app.logger.GetSink() == nil {
		return logr.Discard()
	}
	return app.logger
}

func (app *fakeApp) GetConfig(cfg any) error {
	if app.config == "" {
		return nil
	}
	return yaml.Unmarshal([]byte(app.config), cfg)
}

func (app *fakeApp) RegisterAssetProvider(provider plugin.AssetProvider) error {
	app.assets = provider
	return nil
}

func (app *fakeApp) RegisterQuoteProvider(provider plugin.QuoteProvider) error {
	app.quotes = provider
	return nil
}

func TestPluginInit(t *testing.T) {
	ctx := context.Background()

	t.Run("without an api key registers no providers", func(t *testing.T) {
		app := &fakeApp{}

		require.NoError(t, NewPlugin().Init(app))

		assert.Nil(t, app.assets)
		assert.Nil(t, app.quotes)
	})

	t.Run("with an api key registers asset and quote providers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "test-key", r.Header.Get(demoKeyHeader))
			_, _ = w.Write([]byte(`[{"id":"bitcoin","symbol":"btc","name":"Bitcoin"}]`))
		}))
		t.Cleanup(server.Close)

		app := &fakeApp{config: "api_key: test-key\nbase_url: " + server.URL}

		require.NoError(t, NewPlugin().Init(app))

		require.NotNil(t, app.assets)
		require.NotNil(t, app.quotes)

		assets, err := app.assets.Provide(ctx)

		require.NoError(t, err)
		assert.Equal(t, []plugin.Asset{{ID: "bitcoin", Symbol: "btc", Name: "Bitcoin"}}, assets)
	})
}

func TestPluginKeepsTheAPIKeyOutOfLogs(t *testing.T) {
	const apiKey = "super-secret-key"

	var logs []string
	app := &fakeApp{
		logger: funcr.New(func(_, args string) { logs = append(logs, args) }, funcr.Options{}),
	}

	var gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get(demoKeyHeader)
		_, _ = w.Write([]byte(`[{"id":"bitcoin","symbol":"btc","name":"Bitcoin"}]`))
	}))
	t.Cleanup(server.Close)
	app.config = "api_key: " + apiKey + "\nbase_url: " + server.URL

	require.NoError(t, NewPlugin().Init(app))
	require.NotNil(t, app.quotes)

	assets, err := app.assets.Provide(context.Background())

	require.NoError(t, err)
	assert.NotEmpty(t, assets)
	assert.Equal(t, apiKey, gotKey)

	joined := strings.Join(logs, "\n")
	assert.Contains(t, joined, "/coins/list")
	assert.NotContains(t, joined, apiKey)
}
