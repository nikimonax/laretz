package usecase

import (
	"context"
	"time"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

// GetPortfolioChart requests the hourly value series of one Portfolio.
type GetPortfolioChart struct {
	ID domain.PortfolioID
}

// GetPortfolioChartUseCase computes the hourly value series of one Portfolio
// for display: every hour of the holdings' active intervals carries the
// value of the holdings at that hour, with carried-forward prices flagged.
type GetPortfolioChartUseCase struct {
	portfolios interfaces.PortfolioRepository
	operations interfaces.OperationRepository
	assets     interfaces.AssetRepository
	quotes     interfaces.QuoteRepository
}

func NewGetPortfolioChartUseCase(
	portfolios interfaces.PortfolioRepository,
	operations interfaces.OperationRepository,
	assets interfaces.AssetRepository,
	quotes interfaces.QuoteRepository,
) *GetPortfolioChartUseCase {
	return &GetPortfolioChartUseCase{
		portfolios: portfolios,
		operations: operations,
		assets:     assets,
		quotes:     quotes,
	}
}

func (usecase *GetPortfolioChartUseCase) Execute(ctx context.Context, cmd GetPortfolioChart) ([]domain.HourValue, error) {
	if _, err := usecase.portfolios.Get(ctx, cmd.ID); err != nil {
		return nil, err
	}

	operations, err := usecase.operations.List(ctx, cmd.ID)

	if err != nil {
		return nil, err
	}

	assets, err := usecase.assets.ListByPortfolioID(ctx, cmd.ID)

	if err != nil {
		return nil, err
	}

	valuation, err := buildValuation(ctx, newPriceHistories(usecase.quotes), operations, assets)

	if err != nil {
		return nil, err
	}

	return valuation.Series(time.Now()), nil
}
