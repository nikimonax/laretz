package plugin_test

import (
	"context"

	"github.com/goccy/go-yaml"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/pkg/plugin"
)

var providerStub = interfaces.ExternalAssetProviderFunc(func(context.Context) ([]domain.ExternalAsset, error) {
	return nil, nil
})

var quoteFeedStub = interfaces.QuoteFeedFunc(func(context.Context, []domain.QuoteRange) ([]domain.FetchedQuote, error) {
	return nil, nil
})

type fakeAssetProvider struct {
	assets []plugin.Asset
	err    error
	calls  int
}

func (p *fakeAssetProvider) Provide(context.Context) ([]plugin.Asset, error) {
	p.calls++
	return p.assets, p.err
}

type fakeQuoteFeed struct {
	ranges []plugin.QuoteRange
	quotes []plugin.Quote
	err    error
}

func (f *fakeQuoteFeed) Fetch(_ context.Context, ranges []plugin.QuoteRange) ([]plugin.Quote, error) {
	f.ranges = ranges
	return f.quotes, f.err
}

type yamlConfig string

func (c yamlConfig) Unmarshal(v any) error {
	return yaml.Unmarshal([]byte(c), v)
}
