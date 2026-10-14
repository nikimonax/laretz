package interfaces

import (
	"context"

	"github.com/nikimonax/laretz/internal/domain"
)

type ExternalAssetProvider interface {
	Provide(ctx context.Context) ([]domain.ExternalAsset, error)
}

type ExternalAssetProviderFunc func(ctx context.Context) ([]domain.ExternalAsset, error)

func (f ExternalAssetProviderFunc) Provide(ctx context.Context) ([]domain.ExternalAsset, error) {
	return f(ctx)
}

type ExternalAssetProviderRegistry interface {
	Find(sourceID domain.SourceID) (ExternalAssetProvider, error)
	List() []domain.SourceID
}

type AssetRepository interface {
	Get(ctx context.Context, id domain.AssetID) (domain.Asset, error)
	List(ctx context.Context) ([]domain.Asset, error)
	ListByPortfolioID(ctx context.Context, portfolioID domain.PortfolioID) ([]domain.Asset, error)
	Create(ctx context.Context, externalAssetID domain.ExternalAssetID, symbol string, name string) (domain.AssetID, error)
	Rename(ctx context.Context, id domain.AssetID, symbol string, name string) error
	Relink(ctx context.Context, id domain.AssetID, externalAssetID domain.ExternalAssetID) error
	Delete(ctx context.Context, id domain.AssetID) error
}

type ExternalAssetRepository interface {
	Count(ctx context.Context) (int64, error)
	Get(ctx context.Context, id domain.ExternalAssetID) (domain.ExternalAsset, error)
	Search(ctx context.Context, query string) ([]domain.ExternalAsset, error)
	// Sync persists a catalog dump as-is. Storage owns ID and SyncedAt:
	// entries are merged by (SourceID, ExternalID).
	Sync(ctx context.Context, catalog []domain.ExternalAsset) error
}
