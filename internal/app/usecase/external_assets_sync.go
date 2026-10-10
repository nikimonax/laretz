package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type SyncExternalAssets struct {
	SourceID domain.SourceID
}

type SyncExternalAssetsUseCase struct {
	providers interfaces.ExternalAssetProviderRegistry
	assets    interfaces.ExternalAssetRepository
}

func (usecase *SyncExternalAssetsUseCase) Execute(ctx context.Context, cmd SyncExternalAssets) error {
	provider, err := usecase.providers.Find(cmd.SourceID)

	if err != nil {
		return err
	}

	catalog, err := provider.Provide(ctx)

	if err != nil {
		return err
	}

	return usecase.assets.Sync(ctx, catalog)
}

func NewSyncExternalAssetsUseCase(
	providers interfaces.ExternalAssetProviderRegistry,
	assets interfaces.ExternalAssetRepository,
) *SyncExternalAssetsUseCase {
	return &SyncExternalAssetsUseCase{
		providers: providers,
		assets:    assets,
	}
}
