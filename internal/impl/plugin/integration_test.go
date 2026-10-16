package plugin_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/domain"
	host "github.com/nikimonax/laretz/internal/impl/plugin"
	"github.com/nikimonax/laretz/internal/impl/repository"
	"github.com/nikimonax/laretz/internal/testing/db"
	"github.com/nikimonax/laretz/plugin/coingecko"
)

func newCoingeckoHost(t *testing.T, apiKey string) (*host.Facade, *host.ExternalAssetProviderRegistry, *host.QuoteFeedRegistry) {
	t.Helper()

	plugins := host.NewRegistry()
	require.NoError(t, plugins.Register(coingecko.NewPlugin()))

	providers := host.NewExternalAssetProviderRegistry()
	quotes := host.NewQuoteFeedRegistry()

	var configs map[host.PluginID]host.Unmarshable
	if apiKey != "" {
		configs = map[host.PluginID]host.Unmarshable{
			coingecko.PluginID: yamlConfig(fmt.Sprintf("api_key: %s", apiKey)),
		}
	}

	facade := host.NewFacade(configs, logr.Discard(), *providers, *quotes)
	for _, p := range plugins.GetAll() {
		require.NoError(t, p.Init(facade.WithPlugin(p.ID())))
	}

	return facade, providers, quotes
}

func TestCoingeckoPluginSyncsCatalogIntoSqlite(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		assert.Equal(t, "/coins/list", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("x-cg-demo-api-key"))
		_, _ = w.Write([]byte(`[
			{"id":"bitcoin","symbol":"btc","name":"Bitcoin"},
			{"id":"ethereum","symbol":"eth","name":"Ethereum"}
		]`))
	}))
	t.Cleanup(server.Close)

	plugins := host.NewRegistry()
	require.NoError(t, plugins.Register(coingecko.NewPlugin()))

	providers := host.NewExternalAssetProviderRegistry()
	quotes := host.NewQuoteFeedRegistry()
	configs := map[host.PluginID]host.Unmarshable{
		coingecko.PluginID: yamlConfig(fmt.Sprintf("api_key: test-key\nbase_url: %s", server.URL)),
	}
	facade := host.NewFacade(configs, logr.Discard(), *providers, *quotes)
	for _, p := range plugins.GetAll() {
		require.NoError(t, p.Init(facade.WithPlugin(p.ID())))
	}

	assets := repository.NewSqliteExternalAssetRepository(db.New(t))
	ctx := context.Background()
	sync := usecase.NewSyncExternalAssetsUseCase(providers, assets)

	require.NoError(t, sync.Execute(ctx, usecase.SyncExternalAssets{SourceID: coingecko.PluginID}))

	count, err := assets.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	found, err := assets.Search(ctx, "bitcoin")
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, domain.SourceID(coingecko.PluginID), found[0].SourceID)
	assert.Equal(t, "bitcoin", found[0].ExternalID)
	assert.Equal(t, "btc", found[0].Symbol)
	assert.Equal(t, "Bitcoin", found[0].Name)
	assert.False(t, found[0].SyncedAt.IsZero())

	require.NoError(t, sync.Execute(ctx, usecase.SyncExternalAssets{SourceID: coingecko.PluginID}))
	assert.Equal(t, 2, requests)

	count, err = assets.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestCoingeckoPluginWithoutAPIKeyRegistersNothing(t *testing.T) {
	_, providers, quotes := newCoingeckoHost(t, "")

	assert.Empty(t, providers.List())

	_, err := quotes.Find(coingecko.PluginID)
	require.ErrorIs(t, err, host.ErrQuoteFeedNotFound)
}
