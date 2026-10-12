package usecase

import (
	"context"
	"time"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

const operationsPageSize = 10

type GetPortfolio struct {
	ID      domain.PortfolioID
	OpsPage int
}

type OperationsPage struct {
	Items []domain.Operation
	Page  int
	Pages int
	Total int64
}

type PortfolioDetail struct {
	Portfolio  domain.Portfolio
	Total      domain.Value
	Positions  []domain.ValuedPosition
	Operations OperationsPage
}

type GetPortfolioUseCase struct {
	portfolios interfaces.PortfolioRepository
	operations interfaces.OperationRepository
	assets     interfaces.AssetRepository
	quotes     interfaces.QuoteRepository
}

func NewGetPortfolioUseCase(
	portfolios interfaces.PortfolioRepository,
	operations interfaces.OperationRepository,
	assets interfaces.AssetRepository,
	quotes interfaces.QuoteRepository,
) *GetPortfolioUseCase {
	return &GetPortfolioUseCase{
		portfolios: portfolios,
		operations: operations,
		assets:     assets,
		quotes:     quotes,
	}
}

func (usecase *GetPortfolioUseCase) Execute(ctx context.Context, cmd GetPortfolio) (PortfolioDetail, error) {
	portfolio, err := usecase.portfolios.Get(ctx, cmd.ID)

	if err != nil {
		return PortfolioDetail{}, err
	}

	operations, err := usecase.operations.List(ctx, cmd.ID)

	if err != nil {
		return PortfolioDetail{}, err
	}

	assets, err := usecase.assets.List(ctx)

	if err != nil {
		return PortfolioDetail{}, err
	}

	histories := newPriceHistories(usecase.quotes)

	valuation, err := buildValuation(ctx, histories, operations, assets)

	if err != nil {
		return PortfolioDetail{}, err
	}

	total, positions := valuation.At(time.Now())

	totalCount, err := usecase.operations.Count(ctx, cmd.ID)

	if err != nil {
		return PortfolioDetail{}, err
	}

	pages := int((totalCount + operationsPageSize - 1) / operationsPageSize)

	if pages == 0 {
		pages = 1
	}

	page := cmd.OpsPage

	if page < 1 {
		page = 1
	}

	if page > pages {
		page = pages
	}

	items, err := usecase.operations.ListPaged(ctx, cmd.ID, operationsPageSize, (page-1)*operationsPageSize)

	if err != nil {
		return PortfolioDetail{}, err
	}

	return PortfolioDetail{
		Portfolio: portfolio,
		Total:     total,
		Positions: positions,
		Operations: OperationsPage{
			Items: items,
			Page:  page,
			Pages: pages,
			Total: totalCount,
		},
	}, nil
}
