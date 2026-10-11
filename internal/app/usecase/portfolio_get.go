package usecase

import (
	"context"

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
	Operations OperationsPage
}

type GetPortfolioUseCase struct {
	portfolios interfaces.PortfolioRepository
	operations interfaces.OperationRepository
}

func NewGetPortfolioUseCase(portfolios interfaces.PortfolioRepository, operations interfaces.OperationRepository) *GetPortfolioUseCase {
	return &GetPortfolioUseCase{portfolios: portfolios, operations: operations}
}

func (usecase *GetPortfolioUseCase) Execute(ctx context.Context, cmd GetPortfolio) (PortfolioDetail, error) {
	portfolio, err := usecase.portfolios.Get(ctx, cmd.ID)

	if err != nil {
		return PortfolioDetail{}, err
	}

	total, err := usecase.operations.Count(ctx, cmd.ID)

	if err != nil {
		return PortfolioDetail{}, err
	}

	pages := int((total + operationsPageSize - 1) / operationsPageSize)

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

	items, err := usecase.operations.List(ctx, cmd.ID, operationsPageSize, (page-1)*operationsPageSize)

	if err != nil {
		return PortfolioDetail{}, err
	}

	return PortfolioDetail{
		Portfolio: portfolio,
		Operations: OperationsPage{
			Items: items,
			Page:  page,
			Pages: pages,
			Total: total,
		},
	}, nil
}
