package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type DeleteAsset struct {
	ID domain.AssetID
}

type DeleteAssetUseCase struct {
	assets interfaces.AssetRepository
}

func (usecase *DeleteAssetUseCase) Execute(ctx context.Context, cmd DeleteAsset) error {
	return usecase.assets.Delete(ctx, cmd.ID)
}

func NewDeleteAssetUseCase(assets interfaces.AssetRepository) *DeleteAssetUseCase {
	return &DeleteAssetUseCase{assets: assets}
}
