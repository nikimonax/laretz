package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/impl/repository"
	"github.com/nikimonax/laretz/internal/testing/db"
)

func TestQuoteRepository(t *testing.T) {
	ctx := context.Background()
	database := db.New(t)

	externalAssets := repository.NewSqliteExternalAssetRepository(database)
	quotes := repository.NewSqliteQuoteRepository(database)

	btc := seedExternalAsset(ctx, t, externalAssets, sourceCoingecko, "bitcoin", "BTC")
	eth := seedExternalAsset(ctx, t, externalAssets, sourceCoingecko, "ethereum", "ETH")

	hour := func(day int, hour int) time.Time {
		return time.Date(2026, 3, day, hour, 0, 0, 0, time.UTC)
	}

	t.Run("store and coverage round trip ordered by hour", func(t *testing.T) {
		require.NoError(t, quotes.Store(ctx, []domain.PricePoint{
			{ExternalAssetID: btc.ID, At: hour(2, 11), PriceUSD: decimal.RequireFromString("11.00"), Granularity: domain.GranularityHour},
			{ExternalAssetID: btc.ID, At: hour(2, 10), PriceUSD: decimal.RequireFromString("10.50"), Granularity: domain.GranularityHour},
			{ExternalAssetID: eth.ID, At: hour(2, 10), PriceUSD: decimal.RequireFromString("100.25"), Granularity: domain.GranularityDay},
		}))

		points, err := quotes.Coverage(ctx, btc.ID)
		require.NoError(t, err)
		require.Len(t, points, 2)

		assert.Equal(t, hour(2, 10), points[0].At)
		assert.Equal(t, btc.ID, points[0].ExternalAssetID)
		assert.True(t, points[0].PriceUSD.Equal(decimal.RequireFromString("10.50")))
		assert.Equal(t, domain.GranularityHour, points[0].Granularity)

		assert.Equal(t, hour(2, 11), points[1].At)
		assert.True(t, points[1].PriceUSD.Equal(decimal.RequireFromString("11.00")))

		points, err = quotes.Coverage(ctx, eth.ID)
		require.NoError(t, err)
		require.Len(t, points, 1)
		assert.Equal(t, domain.GranularityDay, points[0].Granularity)
		assert.True(t, points[0].PriceUSD.Equal(decimal.RequireFromString("100.25")))

		points, err = quotes.Coverage(ctx, 424242)
		require.NoError(t, err)
		assert.Empty(t, points)
	})

	t.Run("store never overwrites existing points", func(t *testing.T) {
		sol := seedExternalAsset(ctx, t, externalAssets, sourceCoingecko, "solana", "SOL")

		require.NoError(t, quotes.Store(ctx, []domain.PricePoint{
			{ExternalAssetID: sol.ID, At: hour(3, 10), PriceUSD: decimal.RequireFromString("10.50"), Granularity: domain.GranularityHour},
		}))

		require.NoError(t, quotes.Store(ctx, []domain.PricePoint{
			{ExternalAssetID: sol.ID, At: hour(3, 10), PriceUSD: decimal.RequireFromString("99.99"), Granularity: domain.GranularityHour},
			{ExternalAssetID: sol.ID, At: hour(3, 12), PriceUSD: decimal.RequireFromString("12.00"), Granularity: domain.GranularityHour},
		}))

		points, err := quotes.Coverage(ctx, sol.ID)
		require.NoError(t, err)
		require.Len(t, points, 2)

		assert.Equal(t, hour(3, 10), points[0].At)
		assert.True(t, points[0].PriceUSD.Equal(decimal.RequireFromString("10.50")))

		assert.Equal(t, hour(3, 12), points[1].At)
		assert.True(t, points[1].PriceUSD.Equal(decimal.RequireFromString("12.00")))
	})
}
