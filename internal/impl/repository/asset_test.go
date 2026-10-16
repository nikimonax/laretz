package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/impl/repository"
	"github.com/nikimonax/laretz/internal/testing/db"
)

func TestAssetRepository(t *testing.T) {
	ctx := context.Background()
	database := db.New(t)

	portfolios := repository.NewSqlitePortfolioRepository(database)
	externalAssets := repository.NewSqliteExternalAssetRepository(database)
	assets := repository.NewSqliteAssetRepository(database)
	operations := repository.NewSqliteOperationRepository(database)

	coingeckoBTC := seedExternalAsset(ctx, t, externalAssets, sourceCoingecko, "bitcoin", "BTC")
	coingeckoETH := seedExternalAsset(ctx, t, externalAssets, sourceCoingecko, "ethereum", "ETH")
	binanceBTC := seedExternalAsset(ctx, t, externalAssets, sourceBinance, "BTCUSDT", "BTC")

	t.Run("create requires an existing external asset", func(t *testing.T) {
		_, err := assets.Create(ctx, 424242, "ORPHAN", "Orphan")
		require.ErrorIs(t, err, app.ErrEntityNotFound)

		list, err := assets.List(ctx)
		require.NoError(t, err)

		for _, asset := range list {
			assert.NotEqual(t, "ORPHAN", asset.Symbol)
		}
	})

	t.Run("create get and list round trip joins the external asset", func(t *testing.T) {
		eth := seedAsset(ctx, t, assets, coingeckoETH, "ETH")
		btc := seedAsset(ctx, t, assets, coingeckoBTC, "BTC")

		got, err := assets.Get(ctx, btc.ID)
		require.NoError(t, err)

		assert.Equal(t, btc.ID, got.ID)
		assert.Equal(t, "BTC", got.Symbol)
		assert.Equal(t, "BTC name", got.Name)
		assert.False(t, got.CreatedAt.IsZero())
		assert.False(t, got.UpdatedAt.IsZero())

		assert.Equal(t, coingeckoBTC.ID, got.External.ID)
		assert.Equal(t, sourceCoingecko, got.External.SourceID)
		assert.Equal(t, "bitcoin", got.External.ExternalID)
		assert.Equal(t, "BTC", got.External.Symbol)
		assert.False(t, got.External.SyncedAt.IsZero())

		list, err := assets.List(ctx)
		require.NoError(t, err)

		var btcIndex, ethIndex int

		for index, asset := range list {
			switch asset.ID {
			case btc.ID:
				btcIndex = index
			case eth.ID:
				ethIndex = index
			}
		}

		assert.Less(t, btcIndex, ethIndex)
	})

	t.Run("rename updates fields in place", func(t *testing.T) {
		asset := seedAsset(ctx, t, assets, coingeckoBTC, "REN")

		require.NoError(t, assets.Rename(ctx, asset.ID, "RENN", "Renamed"))
		got, err := assets.Get(ctx, asset.ID)
		require.NoError(t, err)
		assert.Equal(t, "RENN", got.Symbol)
		assert.Equal(t, "Renamed", got.Name)
		assert.False(t, got.UpdatedAt.Before(got.CreatedAt))

		require.ErrorIs(t, assets.Rename(ctx, 424242, "X", "Y"), app.ErrEntityNotFound)
	})

	t.Run("relink swaps the catalog entry within the same source", func(t *testing.T) {
		asset := seedAsset(ctx, t, assets, coingeckoBTC, "RL")

		require.NoError(t, assets.Relink(ctx, asset.ID, coingeckoETH.ID))
		got, err := assets.Get(ctx, asset.ID)
		require.NoError(t, err)
		assert.Equal(t, coingeckoETH.ID, got.External.ID)
		assert.Equal(t, sourceCoingecko, got.External.SourceID)

		err = assets.Relink(ctx, asset.ID, binanceBTC.ID)
		require.ErrorIs(t, err, app.ErrSourceMismatch)

		got, err = assets.Get(ctx, asset.ID)
		require.NoError(t, err)
		assert.Equal(t, coingeckoETH.ID, got.External.ID)

		require.ErrorIs(t, assets.Relink(ctx, asset.ID, 424242), app.ErrEntityNotFound)
		require.ErrorIs(t, assets.Relink(ctx, 424242, coingeckoBTC.ID), app.ErrEntityNotFound)
	})

	t.Run("delete removes the asset", func(t *testing.T) {
		asset := seedAsset(ctx, t, assets, coingeckoBTC, "DEL")

		require.NoError(t, assets.Delete(ctx, asset.ID))

		_, err := assets.Get(ctx, asset.ID)
		require.ErrorIs(t, err, app.ErrEntityNotFound)
		require.ErrorIs(t, assets.Delete(ctx, asset.ID), app.ErrEntityNotFound)
	})

	t.Run("delete is blocked while operations exist", func(t *testing.T) {
		portfolio := seedPortfolio(ctx, t, portfolios)
		asset := seedAsset(ctx, t, assets, coingeckoBTC, "HOLD")

		require.NoError(t, operations.Create(ctx, domain.Operation{
			PortfolioID: portfolio.ID,
			AssetID:     asset.ID,
			Type:        domain.OperationBuy,
			ExecutedAt:  time.Date(2026, 1, 2, 13, 0, 0, 0, time.UTC),
			Quantity:    decimal.RequireFromString("1"),
			TotalUSD:    decimal.RequireFromString("10"),
		}))

		require.Error(t, assets.Delete(ctx, asset.ID))

		got, err := assets.Get(ctx, asset.ID)
		require.NoError(t, err)
		assert.Equal(t, asset.ID, got.ID)
	})
}
