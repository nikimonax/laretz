package plugin_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/domain"
	host "github.com/nikimonax/laretz/internal/impl/plugin"
	"github.com/nikimonax/laretz/pkg/plugin"
)

type fakePlugin struct {
	id string
}

func (p *fakePlugin) ID() string { return p.id }

func (p *fakePlugin) Init(plugin.App) error { return nil }

func TestRegistry(t *testing.T) {
	t.Run("registers and returns plugins", func(t *testing.T) {
		registry := host.NewRegistry()
		first := &fakePlugin{id: "coingecko"}
		second := &fakePlugin{id: "binance"}

		require.NoError(t, registry.Register(first))
		require.NoError(t, registry.Register(second))

		found, err := registry.Get("binance")
		require.NoError(t, err)
		assert.Equal(t, "binance", found.ID())

		registered := registry.GetAll()
		require.Len(t, registered, 2)
		ids := make([]string, 0, len(registered))
		for _, p := range registered {
			ids = append(ids, p.ID())
		}
		assert.ElementsMatch(t, []string{"coingecko", "binance"}, ids)
	})

	t.Run("rejects a duplicate plugin", func(t *testing.T) {
		registry := host.NewRegistry()
		require.NoError(t, registry.Register(&fakePlugin{id: "coingecko"}))

		err := registry.Register(&fakePlugin{id: "coingecko"})

		require.ErrorIs(t, err, host.ErrPluginAlreadyRegistered)
		assert.ErrorContains(t, err, "coingecko")
	})

	t.Run("get missing returns an error", func(t *testing.T) {
		registry := host.NewRegistry()

		_, err := registry.Get("binance")

		require.ErrorIs(t, err, host.ErrPluginNotFound)
		assert.ErrorContains(t, err, "binance")
	})
}

func TestExternalAssetProviderRegistry(t *testing.T) {
	t.Run("lists the registered sources sorted", func(t *testing.T) {
		registry := host.NewExternalAssetProviderRegistry()

		require.NoError(t, registry.Register("coingecko", providerStub))
		require.NoError(t, registry.Register("binance", providerStub))

		assert.Equal(t, []domain.SourceID{"binance", "coingecko"}, registry.List())
	})

	t.Run("find returns the registered provider", func(t *testing.T) {
		registry := host.NewExternalAssetProviderRegistry()

		require.NoError(t, registry.Register("coingecko", providerStub))

		found, err := registry.Find("coingecko")
		require.NoError(t, err)
		assert.NotNil(t, found)
	})

	t.Run("rejects a duplicate source", func(t *testing.T) {
		registry := host.NewExternalAssetProviderRegistry()

		require.NoError(t, registry.Register("coingecko", providerStub))

		err := registry.Register("coingecko", providerStub)

		require.ErrorIs(t, err, host.ErrAssetProviderAlreadyRegistered)
	})

	t.Run("find missing returns an error", func(t *testing.T) {
		registry := host.NewExternalAssetProviderRegistry()

		_, err := registry.Find("binance")

		require.ErrorIs(t, err, host.ErrAssetProviderNotFound)
	})
}

func TestQuoteFeedRegistry(t *testing.T) {
	t.Run("find returns the registered feed", func(t *testing.T) {
		registry := host.NewQuoteFeedRegistry()

		require.NoError(t, registry.Register("coingecko", quoteFeedStub))

		found, err := registry.Find("coingecko")
		require.NoError(t, err)
		assert.NotNil(t, found)
	})

	t.Run("rejects a duplicate source", func(t *testing.T) {
		registry := host.NewQuoteFeedRegistry()

		require.NoError(t, registry.Register("coingecko", quoteFeedStub))

		err := registry.Register("coingecko", quoteFeedStub)

		require.ErrorIs(t, err, host.ErrQuoteFeedAlreadyRegistered)
	})

	t.Run("find missing returns an error", func(t *testing.T) {
		registry := host.NewQuoteFeedRegistry()

		_, err := registry.Find("binance")

		require.ErrorIs(t, err, host.ErrQuoteFeedNotFound)
	})
}
