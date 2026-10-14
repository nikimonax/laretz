package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/testing/mocks"
)

func quotePoint(externalID domain.ExternalAssetID, at time.Time, price string) domain.PricePoint {
	return domain.PricePoint{
		ExternalAssetID: externalID,
		At:              at,
		PriceUSD:        decimal.RequireFromString(price),
		Granularity:     domain.GranularityHour,
	}
}

func fetchedQuote(source domain.SourceID, externalID string, at time.Time, price string, granularity domain.Granularity) domain.FetchedQuote {
	return domain.FetchedQuote{
		SourceID:    source,
		ExternalID:  externalID,
		At:          at,
		PriceUSD:    decimal.RequireFromString(price),
		Granularity: granularity,
	}
}

func TestGetPortfolioChartUseCase(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Hour)
	from := now.Add(-2 * time.Hour)

	t.Run("builds the hourly series with carried-forward prices", func(t *testing.T) {
		portfolios := &mocks.MoqPortfolioRepository{}
		portfolios.GetFunc = func(context.Context, domain.PortfolioID) (domain.Portfolio, error) {
			return domain.Portfolio{ID: 1, Name: "Main"}, nil
		}

		operations := &mocks.MoqOperationRepository{}
		operations.ListFunc = func(context.Context, domain.PortfolioID) ([]domain.Operation, error) {
			return []domain.Operation{testOperation(1, 1, 1, domain.OperationBuy, from, "2")}, nil
		}

		assets := &mocks.MoqAssetRepository{}
		assets.ListByPortfolioIDFunc = func(context.Context, domain.PortfolioID) ([]domain.Asset, error) {
			return []domain.Asset{testAsset(1, "BTC", 10)}, nil
		}

		quotes := &mocks.MoqQuoteRepository{}
		quotes.CoverageFunc = func(context.Context, domain.ExternalAssetID) ([]domain.PricePoint, error) {
			return []domain.PricePoint{quotePoint(10, from, "100")}, nil
		}

		series, err := usecase.NewGetPortfolioChartUseCase(portfolios, operations, assets, quotes).
			Execute(ctx, usecase.GetPortfolioChart{ID: 1})

		require.NoError(t, err)
		assert.Len(t, series, 3)
		assert.Equal(t, from, series[0].Hour)
		assert.True(t, series[0].Held)
		assert.True(t, decimal.RequireFromString("200").Equal(series[0].Amount))
		assert.False(t, series[0].Stale)
		assert.True(t, series[1].Stale)
		assert.True(t, series[2].Stale)
	})

	t.Run("portfolio with no operations yields no series", func(t *testing.T) {
		portfolios := &mocks.MoqPortfolioRepository{}
		portfolios.GetFunc = func(context.Context, domain.PortfolioID) (domain.Portfolio, error) {
			return domain.Portfolio{ID: 1}, nil
		}

		operations := &mocks.MoqOperationRepository{}
		operations.ListFunc = func(context.Context, domain.PortfolioID) ([]domain.Operation, error) { return nil, nil }

		assets := &mocks.MoqAssetRepository{}
		assets.ListByPortfolioIDFunc = func(context.Context, domain.PortfolioID) ([]domain.Asset, error) { return nil, nil }

		quotes := &mocks.MoqQuoteRepository{}

		series, err := usecase.NewGetPortfolioChartUseCase(portfolios, operations, assets, quotes).
			Execute(ctx, usecase.GetPortfolioChart{ID: 1})

		require.NoError(t, err)
		assert.Nil(t, series)
		assert.Empty(t, quotes.CoverageCalls(), "no prices requested for an empty portfolio")
	})

	t.Run("unknown portfolio fails", func(t *testing.T) {
		portfolios := &mocks.MoqPortfolioRepository{}
		portfolios.GetFunc = func(context.Context, domain.PortfolioID) (domain.Portfolio, error) {
			return domain.Portfolio{}, app.ErrEntityNotFound
		}

		_, err := usecase.NewGetPortfolioChartUseCase(portfolios, &mocks.MoqOperationRepository{},
			&mocks.MoqAssetRepository{}, &mocks.MoqQuoteRepository{}).Execute(ctx, usecase.GetPortfolioChart{ID: 999})

		require.ErrorIs(t, err, app.ErrEntityNotFound)
	})
}

func TestFetchQuotesUseCase(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Hour)
	from := now.Add(-2 * time.Hour)

	type fixture struct {
		quotes *mocks.MoqQuoteRepository
		feed   *mocks.MoqQuoteFeed
		uc     *usecase.FetchQuotesUseCase
	}

	setup := func(coverage []domain.PricePoint, fetched []domain.FetchedQuote) *fixture {
		portfolios := &mocks.MoqPortfolioRepository{}
		portfolios.GetFunc = func(context.Context, domain.PortfolioID) (domain.Portfolio, error) {
			return domain.Portfolio{ID: 1}, nil
		}

		operations := &mocks.MoqOperationRepository{}
		operations.ListFunc = func(context.Context, domain.PortfolioID) ([]domain.Operation, error) {
			return []domain.Operation{testOperation(1, 1, 1, domain.OperationBuy, from, "1")}, nil
		}

		assets := &mocks.MoqAssetRepository{}
		assets.ListByPortfolioIDFunc = func(context.Context, domain.PortfolioID) ([]domain.Asset, error) {
			return []domain.Asset{testAsset(1, "BTC", 10)}, nil
		}

		quotes := &mocks.MoqQuoteRepository{}
		quotes.CoverageFunc = func(context.Context, domain.ExternalAssetID) ([]domain.PricePoint, error) {
			return coverage, nil
		}
		quotes.StoreFunc = func(_ context.Context, _ []domain.PricePoint) error {
			return nil
		}

		feed := &mocks.MoqQuoteFeed{}
		feed.FetchFunc = func(_ context.Context, _ []domain.QuoteRange) ([]domain.FetchedQuote, error) {
			return fetched, nil
		}

		registry := &mocks.MoqQuoteFeedRegistry{}
		registry.FindFunc = func(domain.SourceID) (interfaces.QuoteFeed, error) { return feed, nil }

		return &fixture{
			quotes: quotes,
			feed:   feed,
			uc:     usecase.NewFetchQuotesUseCase(portfolios, operations, assets, quotes, registry),
		}
	}

	t.Run("fetches only the missing gap and stores the points", func(t *testing.T) {
		// The first hour is already covered; only the next two are missing.
		f := setup(
			[]domain.PricePoint{quotePoint(10, from, "100")},
			[]domain.FetchedQuote{
				fetchedQuote("coingecko", "BTC", from.Add(time.Hour), "110", domain.GranularityHour),
				fetchedQuote("coingecko", "BTC", now, "120", domain.GranularityHour),
			},
		)

		err := f.uc.Execute(ctx, usecase.FetchQuotes{PortfolioID: 1})

		require.NoError(t, err)
		assert.Equal(t, []domain.QuoteRange{{
			SourceID:   "coingecko",
			ExternalID: "BTC",
			From:       from.Add(time.Hour),
			To:         now,
		}}, f.feed.FetchCalls()[0].Ranges)
		assert.Len(t, f.quotes.StoreCalls()[0].Points, 2)
		assert.Equal(t, domain.ExternalAssetID(10), f.quotes.StoreCalls()[0].Points[0].ExternalAssetID)
		assert.True(t, decimal.RequireFromString("110").Equal(f.quotes.StoreCalls()[0].Points[0].PriceUSD))
	})

	t.Run("daily granularity covers the whole day and is stored as daily", func(t *testing.T) {
		dayStart := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)

		f := setup(nil, []domain.FetchedQuote{
			fetchedQuote("coingecko", "BTC", dayStart, "100", domain.GranularityDay),
		})

		err := f.uc.Execute(ctx, usecase.FetchQuotes{PortfolioID: 1})

		require.NoError(t, err)
		assert.Len(t, f.quotes.StoreCalls()[0].Points, 1)
		assert.Equal(t, domain.GranularityDay, f.quotes.StoreCalls()[0].Points[0].Granularity)
	})

	t.Run("asset without a primary link gets no prices", func(t *testing.T) {
		portfolios := &mocks.MoqPortfolioRepository{}
		portfolios.GetFunc = func(context.Context, domain.PortfolioID) (domain.Portfolio, error) {
			return domain.Portfolio{ID: 1}, nil
		}

		operations := &mocks.MoqOperationRepository{}
		operations.ListFunc = func(context.Context, domain.PortfolioID) ([]domain.Operation, error) {
			return []domain.Operation{testOperation(1, 1, 1, domain.OperationBuy, from, "1")}, nil
		}

		assets := &mocks.MoqAssetRepository{}
		assets.ListByPortfolioIDFunc = func(context.Context, domain.PortfolioID) ([]domain.Asset, error) {
			return []domain.Asset{testAsset(1, "BTC", 0)}, nil
		}

		quotes := &mocks.MoqQuoteRepository{}
		feed := &mocks.MoqQuoteFeed{}
		registry := &mocks.MoqQuoteFeedRegistry{}
		registry.FindFunc = func(domain.SourceID) (interfaces.QuoteFeed, error) { return feed, nil }

		err := usecase.NewFetchQuotesUseCase(portfolios, operations, assets, quotes, registry).
			Execute(ctx, usecase.FetchQuotes{PortfolioID: 1})

		require.NoError(t, err)
		assert.Empty(t, feed.FetchCalls(), "no feed request for an unlinked asset")
		assert.Empty(t, quotes.StoreCalls())
	})

	t.Run("unknown portfolio fails", func(t *testing.T) {
		portfolios := &mocks.MoqPortfolioRepository{}
		portfolios.GetFunc = func(context.Context, domain.PortfolioID) (domain.Portfolio, error) {
			return domain.Portfolio{}, app.ErrEntityNotFound
		}

		assets := &mocks.MoqAssetRepository{}
		assets.ListByPortfolioIDFunc = func(context.Context, domain.PortfolioID) ([]domain.Asset, error) { return nil, nil }

		err := usecase.NewFetchQuotesUseCase(portfolios, &mocks.MoqOperationRepository{},
			assets, &mocks.MoqQuoteRepository{}, &mocks.MoqQuoteFeedRegistry{}).
			Execute(ctx, usecase.FetchQuotes{PortfolioID: 999})

		require.ErrorIs(t, err, app.ErrEntityNotFound)
	})
}

func TestGetPortfolioUseCase(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Hour)

	t.Run("returns detail with clamped operations page", func(t *testing.T) {
		portfolios := &mocks.MoqPortfolioRepository{}
		portfolios.GetFunc = func(context.Context, domain.PortfolioID) (domain.Portfolio, error) {
			return domain.Portfolio{ID: 1, Name: "Main"}, nil
		}

		operations := &mocks.MoqOperationRepository{}
		operations.ListFunc = func(context.Context, domain.PortfolioID) ([]domain.Operation, error) {
			return []domain.Operation{testOperation(1, 1, 1, domain.OperationBuy, now.Add(-time.Hour), "1")}, nil
		}
		operations.CountFunc = func(context.Context, domain.PortfolioID) (int64, error) { return 25, nil }
		operations.ListPagedFunc = func(_ context.Context, _ domain.PortfolioID, limit int, offset int) ([]domain.Operation, error) {
			assert.Equal(t, 10, limit)
			assert.Equal(t, 20, offset)

			return []domain.Operation{testOperation(21, 1, 1, domain.OperationBuy, now.Add(-time.Hour), "1")}, nil
		}

		assets := &mocks.MoqAssetRepository{}
		assets.ListByPortfolioIDFunc = func(context.Context, domain.PortfolioID) ([]domain.Asset, error) {
			return []domain.Asset{testAsset(1, "BTC", 10)}, nil
		}

		quotes := &mocks.MoqQuoteRepository{}
		quotes.CoverageFunc = func(context.Context, domain.ExternalAssetID) ([]domain.PricePoint, error) {
			return []domain.PricePoint{quotePoint(10, now.Add(-time.Hour), "100")}, nil
		}

		detail, err := usecase.NewGetPortfolioUseCase(portfolios, operations, assets, quotes).
			Execute(ctx, usecase.GetPortfolio{ID: 1, OpsPage: 99})

		require.NoError(t, err)
		assert.Equal(t, "Main", detail.Portfolio.Name)
		assert.Equal(t, int64(25), detail.Operations.Total)
		assert.Equal(t, 3, detail.Operations.Pages)
		assert.Equal(t, 3, detail.Operations.Page)
		assert.Len(t, detail.Operations.Items, 1)
		assert.Len(t, detail.Positions, 1)
	})

	t.Run("empty portfolio yields an empty first page", func(t *testing.T) {
		portfolios := &mocks.MoqPortfolioRepository{}
		portfolios.GetFunc = func(context.Context, domain.PortfolioID) (domain.Portfolio, error) {
			return domain.Portfolio{ID: 1}, nil
		}

		operations := &mocks.MoqOperationRepository{}
		operations.ListFunc = func(context.Context, domain.PortfolioID) ([]domain.Operation, error) { return nil, nil }
		operations.CountFunc = func(context.Context, domain.PortfolioID) (int64, error) { return 0, nil }
		operations.ListPagedFunc = func(context.Context, domain.PortfolioID, int, int) ([]domain.Operation, error) {
			return nil, nil
		}

		assets := &mocks.MoqAssetRepository{}
		assets.ListByPortfolioIDFunc = func(context.Context, domain.PortfolioID) ([]domain.Asset, error) { return nil, nil }

		detail, err := usecase.NewGetPortfolioUseCase(portfolios, operations, assets, &mocks.MoqQuoteRepository{}).
			Execute(ctx, usecase.GetPortfolio{ID: 1})

		require.NoError(t, err)
		assert.Equal(t, 1, detail.Operations.Pages)
		assert.Equal(t, 1, detail.Operations.Page)
		assert.Empty(t, detail.Operations.Items)
	})
}

func TestListPortfoliosUseCase(t *testing.T) {
	ctx := context.Background()

	t.Run("summaries carry the display-time value", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Hour)

		portfolios := &mocks.MoqPortfolioRepository{}
		portfolios.ListFunc = func(context.Context) ([]domain.Portfolio, error) {
			return []domain.Portfolio{{ID: 1, Name: "Main"}}, nil
		}

		operations := &mocks.MoqOperationRepository{}
		operations.ListFunc = func(context.Context, domain.PortfolioID) ([]domain.Operation, error) {
			return []domain.Operation{testOperation(1, 1, 1, domain.OperationBuy, now.Add(-time.Hour), "2")}, nil
		}

		assets := &mocks.MoqAssetRepository{}
		assets.ListFunc = func(context.Context) ([]domain.Asset, error) {
			return []domain.Asset{testAsset(1, "BTC", 10)}, nil
		}

		quotes := &mocks.MoqQuoteRepository{}
		quotes.CoverageFunc = func(context.Context, domain.ExternalAssetID) ([]domain.PricePoint, error) {
			return []domain.PricePoint{quotePoint(10, now.Add(-time.Hour), "100")}, nil
		}

		summaries, err := usecase.NewListPortfoliosUseCase(portfolios, operations, assets, quotes).Execute(ctx)

		require.NoError(t, err)
		assert.Len(t, summaries, 1)
		assert.True(t, decimal.RequireFromString("200").Equal(summaries[0].Value.Amount))
	})
}
