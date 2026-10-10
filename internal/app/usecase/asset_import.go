package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type ImportAsset struct {
	ExternalAssetID domain.ExternalAssetID
	Symbol          string
	Name            string
}

type ImportAssetUseCase struct {
	assets interfaces.AssetRepository
}

func (usecase *ImportAssetUseCase) Execute(ctx context.Context, cmd ImportAsset) error {
	_, err := usecase.assets.Create(ctx, cmd.ExternalAssetID, cmd.Symbol, cmd.Name)
	return err
}

func NewImportAssetUseCase(assets interfaces.AssetRepository) *ImportAssetUseCase {
	return &ImportAssetUseCase{assets: assets}
}
