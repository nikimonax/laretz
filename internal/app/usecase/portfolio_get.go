package usecase

import (
	"context"
	"sort"
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
	Positions  []domain.Position
	Operations OperationsPage
}

type GetPortfolioUseCase struct {
	portfolios interfaces.PortfolioRepository
	operations interfaces.OperationRepository
	assets     interfaces.AssetRepository
}

func NewGetPortfolioUseCase(
	portfolios interfaces.PortfolioRepository,
	operations interfaces.OperationRepository,
	assets interfaces.AssetRepository,
) *GetPortfolioUseCase {
	return &GetPortfolioUseCase{portfolios: portfolios, operations: operations, assets: assets}
}

func (usecase *GetPortfolioUseCase) Execute(ctx context.Context, cmd GetPortfolio) (PortfolioDetail, error) {
	portfolio, err := usecase.portfolios.Get(ctx, cmd.ID)

	if err != nil {
		return PortfolioDetail{}, err
	}

	positions, err := usecase.positions(ctx, cmd.ID)

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

	items, err := usecase.operations.ListPaged(ctx, cmd.ID, operationsPageSize, (page-1)*operationsPageSize)

	if err != nil {
		return PortfolioDetail{}, err
	}

	return PortfolioDetail{
		Portfolio: portfolio,
		Positions: positions,
		Operations: OperationsPage{
			Items: items,
			Page:  page,
			Pages: pages,
			Total: total,
		},
	}, nil
}

// positions derives Positions from the Operations of the Portfolio: one
// Balance per Asset at the current moment, zero balances excluded.
func (usecase *GetPortfolioUseCase) positions(ctx context.Context, portfolioID domain.PortfolioID) ([]domain.Position, error) {
	operations, err := usecase.operations.List(ctx, portfolioID)

	if err != nil {
		return nil, err
	}

	if len(operations) == 0 {
		return nil, nil
	}

	assets, err := usecase.assets.List(ctx)

	if err != nil {
		return nil, err
	}

	assetByID := make(map[domain.AssetID]domain.Asset, len(assets))

	for _, asset := range assets {
		assetByID[asset.ID] = asset
	}

	operationsByAsset := make(map[domain.AssetID][]domain.Operation)

	for _, operation := range operations {
		operationsByAsset[operation.AssetID] = append(operationsByAsset[operation.AssetID], operation)
	}

	now := time.Now()
	positions := make([]domain.Position, 0, len(operationsByAsset))

	for assetID, assetOperations := range operationsByAsset {
		balance := domain.NewBalanceTimeline(assetOperations).BalanceAt(now)

		if balance.IsZero() {
			continue
		}

		asset, ok := assetByID[assetID]

		if !ok {
			continue
		}

		positions = append(positions, domain.Position{Asset: asset, Balance: balance})
	}

	sort.Slice(positions, func(i, j int) bool {
		if positions[i].Asset.Symbol != positions[j].Asset.Symbol {
			return positions[i].Asset.Symbol < positions[j].Asset.Symbol
		}

		return positions[i].Asset.ID < positions[j].Asset.ID
	})

	return positions, nil
}
