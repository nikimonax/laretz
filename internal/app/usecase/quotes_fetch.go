package usecase

import (
	"context"
	"maps"
	"slices"
	"time"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

// FetchQuotes fetches the PricePoints missing from the QuoteCoverage of the
// ExternalAssets held in one Portfolio.
type FetchQuotes struct {
	PortfolioID domain.PortfolioID
}

type FetchQuotesUseCase struct {
	portfolios interfaces.PortfolioRepository
	operations interfaces.OperationRepository
	assets     interfaces.AssetRepository
	quotes     interfaces.QuoteRepository
	feeds      interfaces.QuoteFeedRegistry
}

func NewFetchQuotesUseCase(
	portfolios interfaces.PortfolioRepository,
	operations interfaces.OperationRepository,
	assets interfaces.AssetRepository,
	quotes interfaces.QuoteRepository,
	feeds interfaces.QuoteFeedRegistry,
) *FetchQuotesUseCase {
	return &FetchQuotesUseCase{
		portfolios: portfolios,
		operations: operations,
		assets:     assets,
		quotes:     quotes,
		feeds:      feeds,
	}
}

func (usecase *FetchQuotesUseCase) Execute(ctx context.Context, cmd FetchQuotes) error {
	assets, err := usecase.assets.ListByPortfolioID(ctx, cmd.PortfolioID)

	if err != nil {
		return err
	}

	required, err := usecase.requiredIntervals(ctx, cmd.PortfolioID)

	if err != nil {
		return err
	}

	ranges, externalIDs, err := usecase.planRanges(ctx, assets, required)

	if err != nil {
		return err
	}

	points, err := usecase.fetchRanges(ctx, ranges, externalIDs)

	if err != nil {
		return err
	}

	if len(points) == 0 {
		return nil
	}

	return usecase.quotes.Store(ctx, points)
}

// planRanges computes the missing QuoteRanges per ExternalAsset: local assets
// linked to the same ExternalAsset contribute to the same coverage plan.
func (usecase *FetchQuotesUseCase) planRanges(
	ctx context.Context,
	assets []domain.Asset,
	required map[domain.AssetID][]domain.HoldingInterval,
) ([]domain.QuoteRange, map[quoteRef]domain.ExternalAssetID, error) {
	intervalsByExternal := make(map[domain.ExternalAssetID][]domain.HoldingInterval)
	externalByID := make(map[domain.ExternalAssetID]domain.ExternalAsset)

	for _, asset := range assets {
		intervals, ok := required[asset.ID]

		// An asset without a primary link or without holding operations
		// gets no prices.
		if !ok || asset.External.ID == 0 {
			continue
		}

		intervalsByExternal[asset.External.ID] = append(intervalsByExternal[asset.External.ID], intervals...)
		externalByID[asset.External.ID] = asset.External
	}

	var (
		ranges      []domain.QuoteRange
		externalIDs = make(map[quoteRef]domain.ExternalAssetID)
	)

	for _, externalID := range slices.Sorted(maps.Keys(intervalsByExternal)) {
		external := externalByID[externalID]

		points, err := usecase.quotes.Coverage(ctx, externalID)

		if err != nil {
			return nil, nil, err
		}

		coverage := domain.NewQuoteCoverage(points)
		ranges = append(ranges, domain.CoveragePlanner{}.Plan(external, intervalsByExternal[externalID], coverage)...)
		externalIDs[quoteRef{external.SourceID, external.ExternalID}] = externalID
	}

	return ranges, externalIDs, nil
}

// requiredIntervals collects the HoldingIntervals of every Asset in the
// Portfolio.
func (usecase *FetchQuotesUseCase) requiredIntervals(ctx context.Context, portfolioID domain.PortfolioID) (map[domain.AssetID][]domain.HoldingInterval, error) {
	if _, err := usecase.portfolios.Get(ctx, portfolioID); err != nil {
		return nil, err
	}

	operations, err := usecase.operations.List(ctx, portfolioID)

	if err != nil {
		return nil, err
	}

	now := time.Now()
	required := make(map[domain.AssetID][]domain.HoldingInterval)

	operationsByAsset := make(map[domain.AssetID][]domain.Operation)

	for _, operation := range operations {
		operationsByAsset[operation.AssetID] = append(operationsByAsset[operation.AssetID], operation)
	}

	for assetID, assetOperations := range operationsByAsset {
		required[assetID] = append(required[assetID], domain.NewBalanceTimeline(assetOperations).HoldingIntervals(now)...)
	}

	return required, nil
}

// fetchRanges requests the ranges from the QuoteFeed of their source and
// resolves the quotes back to ExternalAssets; quotes of unknown assets are
// dropped.
func (usecase *FetchQuotesUseCase) fetchRanges(
	ctx context.Context,
	ranges []domain.QuoteRange,
	externalIDs map[quoteRef]domain.ExternalAssetID,
) ([]domain.PricePoint, error) {
	bySource := groupBySource(ranges)

	var points []domain.PricePoint

	for _, sourceID := range slices.Sorted(maps.Keys(bySource)) {
		feed, err := usecase.feeds.Find(sourceID)

		if err != nil {
			return nil, err
		}

		fetched, err := feed.Fetch(ctx, bySource[sourceID])

		if err != nil {
			return nil, err
		}

		for _, quote := range fetched {
			externalAssetID, ok := externalIDs[quoteRef{quote.SourceID, quote.ExternalID}]

			if !ok {
				continue
			}

			points = append(points, domain.PricePoint{
				ExternalAssetID: externalAssetID,
				At:              quote.At,
				PriceUSD:        quote.PriceUSD,
				Granularity:     quote.Granularity,
			})
		}
	}

	return points, nil
}

type quoteRef struct {
	source   domain.SourceID
	external string
}

func groupBySource(ranges []domain.QuoteRange) map[domain.SourceID][]domain.QuoteRange {
	bySource := make(map[domain.SourceID][]domain.QuoteRange)

	for _, quoteRange := range ranges {
		bySource[quoteRange.SourceID] = append(bySource[quoteRange.SourceID], quoteRange)
	}

	return bySource
}
