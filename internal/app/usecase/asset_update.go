package usecase

import (
	"context"
	"fmt"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type UpdateAsset struct {
	ID              domain.AssetID
	Symbol          *string
	Name            *string
	ExternalAssetID *domain.ExternalAssetID
}

type UpdateAssetUseCase struct {
	assets interfaces.AssetRepository
}

func (usecase *UpdateAssetUseCase) Execute(ctx context.Context, cmd UpdateAsset) error {
	if cmd.Symbol != nil || cmd.Name != nil {
		if cmd.Symbol == nil || cmd.Name == nil {
			return fmt.Errorf("rename asset %d: symbol and name must be renamed together", cmd.ID)
		}

		if err := usecase.assets.Rename(ctx, cmd.ID, *cmd.Symbol, *cmd.Name); err != nil {
			return err
		}
	}

	if cmd.ExternalAssetID != nil {
		if err := usecase.assets.Relink(ctx, cmd.ID, *cmd.ExternalAssetID); err != nil {
			return err
		}
	}

	return nil
}

func NewUpdateAssetUseCase(assets interfaces.AssetRepository) *UpdateAssetUseCase {
	return &UpdateAssetUseCase{assets: assets}
}
