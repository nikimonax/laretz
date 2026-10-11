package interfaces

import (
	"context"

	"github.com/nikimonax/laretz/internal/domain"
)

// QuoteFeed is a port to an external source of prices. The source decides the
// granularity of each range and returns only the data it actually serves;
// gaps are never synthesized.
type QuoteFeed interface {
	Fetch(ctx context.Context, ranges []domain.QuoteRange) ([]domain.FetchedQuote, error)
}

type QuoteFeedFunc func(ctx context.Context, ranges []domain.QuoteRange) ([]domain.FetchedQuote, error)

// Fetch implements [QuoteFeed].
func (f QuoteFeedFunc) Fetch(ctx context.Context, ranges []domain.QuoteRange) ([]domain.FetchedQuote, error) {
	return f(ctx, ranges)
}

type QuoteFeedRegistry interface {
	Find(sourceID domain.SourceID) (QuoteFeed, error)
}

// QuoteRepository is the data-access port for PricePoints.
type QuoteRepository interface {
	// Coverage returns the PricePoints recorded for the ExternalAsset.
	Coverage(ctx context.Context, externalAssetID domain.ExternalAssetID) ([]domain.PricePoint, error)
	// Store records PricePoints exactly as received: existing points are
	// never overwritten, gaps are never synthesized.
	Store(ctx context.Context, points []domain.PricePoint) error
}
