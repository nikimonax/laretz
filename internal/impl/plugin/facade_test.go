package plugin_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/domain"
	host "github.com/nikimonax/laretz/internal/impl/plugin"
	"github.com/nikimonax/laretz/pkg/plugin"
)

func newFacade(configs map[host.PluginID]host.Unmarshable, logger logr.Logger) (*host.Facade, *host.ExternalAssetProviderRegistry, *host.QuoteFeedRegistry) {
	providers := host.NewExternalAssetProviderRegistry()
	quotes := host.NewQuoteFeedRegistry()

	return host.NewFacade(configs, logger, *providers, *quotes), providers, quotes
}

func TestFacadeRegistersAssetProvider(t *testing.T) {
	facade, providers, _ := newFacade(nil, logr.Discard())
	source := &fakeAssetProvider{assets: []plugin.Asset{{ID: "bitcoin", Symbol: "BTC", Name: "Bitcoin"}}}

	require.NoError(t, facade.WithPlugin("coingecko").RegisterAssetProvider(source))

	found, err := providers.Find("coingecko")
	require.NoError(t, err)

	catalog, err := found.Provide(context.Background())
	require.NoError(t, err)

	require.Len(t, catalog, 1)
	assert.Equal(t, domain.SourceID("coingecko"), catalog[0].SourceID)
	assert.Equal(t, "bitcoin", catalog[0].ExternalID)
	assert.Equal(t, "BTC", catalog[0].Symbol)
	assert.Equal(t, "Bitcoin", catalog[0].Name)
}

func TestFacadeRejectsDuplicateAssetProvider(t *testing.T) {
	facade, _, _ := newFacade(nil, logr.Discard())
	source := &fakeAssetProvider{}

	require.NoError(t, facade.WithPlugin("coingecko").RegisterAssetProvider(source))

	err := facade.WithPlugin("coingecko").RegisterAssetProvider(source)

	require.ErrorIs(t, err, host.ErrAssetProviderAlreadyRegistered)
}

func TestFacadePropagatesAssetProviderErrors(t *testing.T) {
	facade, providers, _ := newFacade(nil, logr.Discard())
	failure := assert.AnError
	source := &fakeAssetProvider{err: failure}

	require.NoError(t, facade.WithPlugin("coingecko").RegisterAssetProvider(source))

	found, err := providers.Find("coingecko")
	require.NoError(t, err)

	_, err = found.Provide(context.Background())

	require.ErrorIs(t, err, failure)
}

func TestFacadeRegistersQuoteFeed(t *testing.T) {
	facade, _, quotes := newFacade(nil, logr.Discard())
	at := time.Date(2026, 1, 2, 3, 0, 0, 0, time.UTC)
	source := &fakeQuoteFeed{quotes: []plugin.Quote{{
		ExternalID:  "bitcoin",
		At:          at,
		Price:       "123.45",
		Granularity: plugin.GranularityHour,
	}}}

	require.NoError(t, facade.WithPlugin("coingecko").RegisterQuoteProvider(source))

	found, err := quotes.Find("coingecko")
	require.NoError(t, err)

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	fetched, err := found.Fetch(context.Background(), []domain.QuoteRange{
		{SourceID: "elsewhere", ExternalID: "bitcoin", From: from, To: to},
	})
	require.NoError(t, err)

	require.Len(t, source.ranges, 1)
	assert.Equal(t, []plugin.QuoteRange{{ExternalID: "bitcoin", From: from, To: to}}, source.ranges)

	require.Len(t, fetched, 1)
	assert.Equal(t, domain.SourceID("coingecko"), fetched[0].SourceID)
	assert.Equal(t, "bitcoin", fetched[0].ExternalID)
	assert.Equal(t, at, fetched[0].At)
	assert.True(t, decimal.RequireFromString("123.45").Equal(fetched[0].PriceUSD))
	assert.Equal(t, domain.GranularityHour, fetched[0].Granularity)
}

func TestFacadeRejectsUnparsableQuotePrice(t *testing.T) {
	facade, _, quotes := newFacade(nil, logr.Discard())
	source := &fakeQuoteFeed{quotes: []plugin.Quote{{ExternalID: "bitcoin", Price: "not-a-price"}}}

	require.NoError(t, facade.WithPlugin("coingecko").RegisterQuoteProvider(source))

	found, err := quotes.Find("coingecko")
	require.NoError(t, err)

	_, err = found.Fetch(context.Background(), nil)

	require.Error(t, err)
}

func TestFacadePropagatesQuoteFeedErrors(t *testing.T) {
	facade, _, quotes := newFacade(nil, logr.Discard())
	failure := assert.AnError
	source := &fakeQuoteFeed{err: failure}

	require.NoError(t, facade.WithPlugin("coingecko").RegisterQuoteProvider(source))

	found, err := quotes.Find("coingecko")
	require.NoError(t, err)

	_, err = found.Fetch(context.Background(), nil)

	require.ErrorIs(t, err, failure)
}

func TestFacadeGetConfigUnmarshalsThePluginSection(t *testing.T) {
	configs := map[host.PluginID]host.Unmarshable{
		"coingecko": yamlConfig("api_key: test-key\nbase_url: http://example.com"),
	}
	facade, _, _ := newFacade(configs, logr.Discard())

	var cfg struct {
		APIKey  string `yaml:"api_key"`
		BaseURL string `yaml:"base_url"`
	}
	require.NoError(t, facade.WithPlugin("coingecko").GetConfig(&cfg))
	assert.Equal(t, "test-key", cfg.APIKey)
	assert.Equal(t, "http://example.com", cfg.BaseURL)
}

func TestFacadeGetConfigWithoutSection(t *testing.T) {
	facade, _, _ := newFacade(nil, logr.Discard())

	var cfg struct {
		APIKey string `yaml:"api_key"`
	}
	require.NoError(t, facade.WithPlugin("coingecko").GetConfig(&cfg))
	assert.Empty(t, cfg.APIKey)
}

func TestFacadeScopesLoggerByPlugin(t *testing.T) {
	var logs []string
	logger := funcr.New(func(_, args string) { logs = append(logs, args) }, funcr.Options{})
	facade, _, _ := newFacade(nil, logger)

	facade.WithPlugin("coingecko").Logger().Info("hello")
	facade.WithPlugin("binance").Logger().Info("hello")

	require.Len(t, logs, 2)
	assert.Contains(t, logs[0], "coingecko")
	assert.NotContains(t, logs[0], "binance")
	assert.Contains(t, logs[1], "binance")
	assert.NotContains(t, logs[1], "coingecko")
}
