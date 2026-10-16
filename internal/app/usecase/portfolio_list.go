package usecase

import (
	"context"
	"time"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

// PortfolioSummary is a Portfolio together with its display-time value.
type PortfolioSummary struct {
	Portfolio domain.Portfolio
	Value     domain.Value
}

type ListPortfoliosUseCase struct {
	portfolios interfaces.PortfolioRepository
	operations interfaces.OperationRepository
	assets     interfaces.AssetRepository
	quotes     interfaces.QuoteRepository
}

func NewListPortfoliosUseCase(
	portfolios interfaces.PortfolioRepository,
	operations interfaces.OperationRepository,
	assets interfaces.AssetRepository,
	quotes interfaces.QuoteRepository,
) *ListPortfoliosUseCase {
	return &ListPortfoliosUseCase{
		portfolios: portfolios,
		operations: operations,
		assets:     assets,
		quotes:     quotes,
	}
}

func (usecase *ListPortfoliosUseCase) Execute(ctx context.Context) ([]PortfolioSummary, error) {
	portfolios, err := usecase.portfolios.List(ctx)

	if err != nil {
		return nil, err
	}

	assets, err := usecase.assets.List(ctx)

	if err != nil {
		return nil, err
	}

	histories := newPriceHistories(usecase.quotes)
	now := time.Now()

	summaries := make([]PortfolioSummary, 0, len(portfolios))

	for _, portfolio := range portfolios {
		operations, err := usecase.operations.List(ctx, portfolio.ID)

		if err != nil {
			return nil, err
		}

		valuation, err := buildValuation(ctx, histories, operations, assets)

		if err != nil {
			return nil, err
		}

		value, _ := valuation.At(now)
		summaries = append(summaries, PortfolioSummary{Portfolio: portfolio, Value: value})
	}

	return summaries, nil
}
