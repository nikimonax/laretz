package interfaces

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/dto"
	"github.com/nikimonax/laretz/internal/domain"
)

type ExternalAssetProvider interface {
	Provide(ctx context.Context) ([]dto.ExternalAsset, error)
}

type ExternalAssetProviderFunc func(ctx context.Context) ([]dto.ExternalAsset, error)

func (f ExternalAssetProviderFunc) Provide(ctx context.Context) ([]dto.ExternalAsset, error) {
	return f(ctx)
}

type ExternalAssetProviderRegistry interface {
	Find(sourceID domain.SourceID) (ExternalAssetProvider, error)
	List() []domain.SourceID
}
