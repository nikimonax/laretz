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

func TestOperationRepository(t *testing.T) {
	ctx := context.Background()
	database := db.New(t)

	portfolios := repository.NewSqlitePortfolioRepository(database)
	externalAssets := repository.NewSqliteExternalAssetRepository(database)
	assets := repository.NewSqliteAssetRepository(database)
	operations := repository.NewSqliteOperationRepository(database)

	portfolio := seedPortfolio(ctx, t, portfolios)
	otherPortfolio := seedPortfolio(ctx, t, portfolios)

	btc := seedAsset(ctx, t, assets, seedExternalAsset(ctx, t, externalAssets, sourceCoingecko, "bitcoin", "BTC"), "BTC")
	eth := seedAsset(ctx, t, assets, seedExternalAsset(ctx, t, externalAssets, sourceCoingecko, "ethereum", "ETH"), "ETH")

	at := time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC)

	nextAt := func() time.Time {
		at = at.Add(time.Hour)

		return at
	}

	// newestOperation returns the operation created last in the portfolio.
	newestOperation := func(t *testing.T, portfolioID domain.PortfolioID) domain.Operation {
		t.Helper()

		ops, err := operations.ListPaged(ctx, portfolioID, 1, 0)
		require.NoError(t, err)
		require.Len(t, ops, 1)

		return ops[0]
	}

	t.Run("create and get round trip preserves every field", func(t *testing.T) {
		feeAmount := decimal.RequireFromString("0.25")
		feeAssetID := eth.ID
		executedAt := nextAt().In(time.FixedZone("UTC+2", 2*60*60))

		require.NoError(t, operations.Create(ctx, domain.Operation{
			PortfolioID: portfolio.ID,
			AssetID:     btc.ID,
			Type:        domain.OperationBuy,
			ExecutedAt:  executedAt,
			Quantity:    decimal.RequireFromString("1.5"),
			TotalUSD:    decimal.RequireFromString("10.00"),
			FeeAmount:   &feeAmount,
			FeeAssetID:  &feeAssetID,
			Note:        "lunch money",
		}))

		got := newestOperation(t, portfolio.ID)

		assert.Equal(t, btc.ID, got.AssetID)
		assert.Equal(t, portfolio.ID, got.PortfolioID)
		assert.Equal(t, domain.OperationBuy, got.Type)
		assert.True(t, got.ExecutedAt.Equal(executedAt))
		assert.Equal(t, time.UTC, got.ExecutedAt.Location())
		assert.True(t, got.Quantity.Equal(decimal.RequireFromString("1.5")))
		assert.True(t, got.TotalUSD.Equal(decimal.RequireFromString("10.00")))
		require.NotNil(t, got.FeeAmount)
		assert.True(t, got.FeeAmount.Equal(feeAmount))
		require.NotNil(t, got.FeeAssetID)
		assert.Equal(t, eth.ID, *got.FeeAssetID)
		assert.Equal(t, "lunch money", got.Note)
		assert.False(t, got.CreatedAt.IsZero())
		assert.False(t, got.UpdatedAt.IsZero())
	})

	t.Run("optional fields round trip as empty", func(t *testing.T) {
		require.NoError(t, operations.Create(ctx, domain.Operation{
			PortfolioID: portfolio.ID,
			AssetID:     btc.ID,
			Type:        domain.OperationSell,
			ExecutedAt:  nextAt(),
			Quantity:    decimal.RequireFromString("2"),
			TotalUSD:    decimal.RequireFromString("20"),
		}))

		got := newestOperation(t, portfolio.ID)

		assert.Equal(t, domain.OperationSell, got.Type)
		assert.Nil(t, got.FeeAmount)
		assert.Nil(t, got.FeeAssetID)
		assert.Empty(t, got.Note)
	})

	t.Run("list orders ascending while listpaged orders descending", func(t *testing.T) {
		require.NoError(t, operations.Create(ctx, domain.Operation{
			PortfolioID: otherPortfolio.ID,
			AssetID:     btc.ID,
			Type:        domain.OperationBuy,
			ExecutedAt:  time.Date(2026, 3, 3, 10, 0, 0, 0, time.UTC),
			Quantity:    decimal.RequireFromString("4"),
			TotalUSD:    decimal.RequireFromString("4"),
		}))
		require.NoError(t, operations.Create(ctx, domain.Operation{
			PortfolioID: otherPortfolio.ID,
			AssetID:     btc.ID,
			Type:        domain.OperationBuy,
			ExecutedAt:  time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC),
			Quantity:    decimal.RequireFromString("2"),
			TotalUSD:    decimal.RequireFromString("2"),
		}))
		require.NoError(t, operations.Create(ctx, domain.Operation{
			PortfolioID: otherPortfolio.ID,
			AssetID:     btc.ID,
			Type:        domain.OperationBuy,
			ExecutedAt:  time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
			Quantity:    decimal.RequireFromString("1"),
			TotalUSD:    decimal.RequireFromString("1"),
		}))
		require.NoError(t, operations.Create(ctx, domain.Operation{
			PortfolioID: otherPortfolio.ID,
			AssetID:     btc.ID,
			Type:        domain.OperationBuy,
			ExecutedAt:  time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC),
			Quantity:    decimal.RequireFromString("3"),
			TotalUSD:    decimal.RequireFromString("3"),
		}))

		count, err := operations.Count(ctx, otherPortfolio.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(4), count)

		list, err := operations.List(ctx, otherPortfolio.ID)
		require.NoError(t, err)
		require.Len(t, list, 4)
		// Ties on executed_at break by id ascending: "2" was created before "3".
		assert.Equal(t, []string{"1", "2", "3", "4"}, []string{
			list[0].Quantity.String(),
			list[1].Quantity.String(),
			list[2].Quantity.String(),
			list[3].Quantity.String(),
		})

		// The last page is partial and later offsets are empty.
		page, err := operations.ListPaged(ctx, otherPortfolio.ID, 2, 0)
		require.NoError(t, err)
		require.Len(t, page, 2)
		// Ties on executed_at break by id descending: "3" outranks "2".
		assert.Equal(t, []string{"4", "3"}, []string{page[0].Quantity.String(), page[1].Quantity.String()})

		page, err = operations.ListPaged(ctx, otherPortfolio.ID, 2, 2)
		require.NoError(t, err)
		require.Len(t, page, 2)
		assert.Equal(t, []string{"2", "1"}, []string{page[0].Quantity.String(), page[1].Quantity.String()})

		page, err = operations.ListPaged(ctx, otherPortfolio.ID, 2, 4)
		require.NoError(t, err)
		assert.Empty(t, page)
	})

	t.Run("update replaces every field", func(t *testing.T) {
		require.NoError(t, operations.Create(ctx, domain.Operation{
			PortfolioID: portfolio.ID,
			AssetID:     btc.ID,
			Type:        domain.OperationBuy,
			ExecutedAt:  nextAt(),
			Quantity:    decimal.RequireFromString("10"),
			TotalUSD:    decimal.RequireFromString("100"),
		}))

		op := newestOperation(t, portfolio.ID)

		updated := op
		updated.Type = domain.OperationSell
		updated.ExecutedAt = nextAt()
		updated.Quantity = decimal.RequireFromString("0.4")
		updated.TotalUSD = decimal.RequireFromString("8")
		updated.FeeAmount = nil
		updated.FeeAssetID = nil
		updated.Note = "updated"

		require.NoError(t, operations.Update(ctx, updated))

		got, err := operations.Get(ctx, op.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.OperationSell, got.Type)
		assert.True(t, got.ExecutedAt.Equal(updated.ExecutedAt))
		assert.True(t, got.Quantity.Equal(updated.Quantity))
		assert.True(t, got.TotalUSD.Equal(updated.TotalUSD))
		assert.Nil(t, got.FeeAmount)
		assert.Nil(t, got.FeeAssetID)
		assert.Equal(t, "updated", got.Note)

		updated.ID = 424242
		require.ErrorIs(t, operations.Update(ctx, updated), app.ErrEntityNotFound)
	})

	t.Run("delete removes the operation", func(t *testing.T) {
		require.NoError(t, operations.Create(ctx, domain.Operation{
			PortfolioID: portfolio.ID,
			AssetID:     btc.ID,
			Type:        domain.OperationBuy,
			ExecutedAt:  nextAt(),
			Quantity:    decimal.RequireFromString("5"),
			TotalUSD:    decimal.RequireFromString("50"),
		}))

		op := newestOperation(t, portfolio.ID)

		before, err := operations.Count(ctx, portfolio.ID)
		require.NoError(t, err)

		require.NoError(t, operations.Delete(ctx, op.ID))

		after, err := operations.Count(ctx, portfolio.ID)
		require.NoError(t, err)
		assert.Equal(t, before-1, after)
		_, err = operations.Get(ctx, op.ID)
		require.ErrorIs(t, err, app.ErrEntityNotFound)
		require.ErrorIs(t, operations.Delete(ctx, op.ID), app.ErrEntityNotFound)
	})

	t.Run("create with a missing asset fails and stores nothing", func(t *testing.T) {
		count, err := operations.Count(ctx, portfolio.ID)
		require.NoError(t, err)

		err = operations.Create(ctx, domain.Operation{
			PortfolioID: portfolio.ID,
			AssetID:     424242,
			Type:        domain.OperationBuy,
			ExecutedAt:  nextAt(),
			Quantity:    decimal.RequireFromString("1"),
			TotalUSD:    decimal.RequireFromString("1"),
		})
		require.Error(t, err)

		after, err := operations.Count(ctx, portfolio.ID)
		require.NoError(t, err)
		assert.Equal(t, count, after)
	})
}
