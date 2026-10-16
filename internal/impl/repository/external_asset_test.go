package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/impl/repository"
	"github.com/nikimonax/laretz/internal/testing/db"
)

func TestExternalAssetRepository(t *testing.T) {
	ctx := context.Background()
	externalAssets := repository.NewSqliteExternalAssetRepository(db.New(t))

	t.Run("sync merges the catalog by source and external id", func(t *testing.T) {
		require.NoError(t, externalAssets.Sync(ctx, []domain.ExternalAsset{
			{SourceID: sourceCoingecko, ExternalID: "bitcoin", Symbol: "BTC", Name: "Bitcoin"},
			{SourceID: sourceCoingecko, ExternalID: "ethereum", Symbol: "ETH", Name: "Ethereum"},
		}))

		count, err := externalAssets.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		require.NoError(t, externalAssets.Sync(ctx, []domain.ExternalAsset{
			{SourceID: sourceCoingecko, ExternalID: "bitcoin", Symbol: "WBTC", Name: "Wrapped Bitcoin"},
		}))

		count, err = externalAssets.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		found, err := externalAssets.Search(ctx, "WBTC")
		require.NoError(t, err)
		require.Len(t, found, 1)
		assert.Equal(t, sourceCoingecko, found[0].SourceID)
		assert.Equal(t, "bitcoin", found[0].ExternalID)
		assert.Equal(t, "WBTC", found[0].Symbol)
		assert.Equal(t, "Wrapped Bitcoin", found[0].Name)
		assert.False(t, found[0].SyncedAt.IsZero())
	})

	t.Run("search matches prefixes and ranks the exact symbol first", func(t *testing.T) {
		searchAssets := repository.NewSqliteExternalAssetRepository(db.New(t))

		require.NoError(t, searchAssets.Sync(ctx, []domain.ExternalAsset{
			{SourceID: sourceCoingecko, ExternalID: "wrapped", Symbol: "WRAP", Name: "Wrapped Bitcoin"},
			{SourceID: sourceCoingecko, ExternalID: "wrappeds", Symbol: "WRAPS", Name: "Wrappeds Token"},
			{SourceID: sourceCoingecko, ExternalID: "unrelated", Symbol: "ETH", Name: "Ethereum"},
		}))

		found, err := searchAssets.Search(ctx, "WRAP")
		require.NoError(t, err)
		require.Len(t, found, 2)
		assert.Equal(t, "WRAP", found[0].Symbol)
		assert.Equal(t, "WRAPS", found[1].Symbol)

		found, err = searchAssets.Search(ctx, "Wrappeds")
		require.NoError(t, err)
		require.Len(t, found, 1)
		assert.Equal(t, "WRAPS", found[0].Symbol)

		found, err = searchAssets.Search(ctx, "Wrapped ZZZZ")
		require.NoError(t, err)
		assert.Empty(t, found)

		found, err = searchAssets.Search(ctx, `Wra"oted`)
		require.NoError(t, err)
		assert.Empty(t, found)
	})

	t.Run("search without a matchable query returns nothing", func(t *testing.T) {
		searchAssets := repository.NewSqliteExternalAssetRepository(db.New(t))

		found, err := searchAssets.Search(ctx, "")
		require.NoError(t, err)
		assert.Empty(t, found)

		found, err = searchAssets.Search(ctx, "   ")
		require.NoError(t, err)
		assert.Empty(t, found)
	})

	t.Run("get missing returns ErrEntityNotFound", func(t *testing.T) {
		_, err := externalAssets.Get(ctx, 424242)
		require.ErrorIs(t, err, app.ErrEntityNotFound)
	})
}
