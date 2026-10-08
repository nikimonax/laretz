package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/commands"
	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type SyncExternalAssets struct {
	SourceID domain.SourceID
}

type SyncExternalAssetsUseCase struct {
	providers interfaces.ExternalAssetProviderRegistry
	handler   interfaces.CommandHandler[commands.SyncExternalAssetsCommand]
}

func (usecase *SyncExternalAssetsUseCase) Execute(ctx context.Context, cmd SyncExternalAssets) error {
	provider, err := usecase.providers.Find(cmd.SourceID)

	if err != nil {
		return err
	}

	assets, err := provider.Provide(ctx)

	if err != nil {
		return err
	}

	cmdNext := commands.SyncExternalAssetsCommand{Assets: assets}
	return usecase.handler.Handle(ctx, cmdNext)
}

func NewSyncExternalAssetsUseCase(
	providers interfaces.ExternalAssetProviderRegistry,
	handler interfaces.CommandHandler[commands.SyncExternalAssetsCommand],
) *SyncExternalAssetsUseCase {
	return &SyncExternalAssetsUseCase{
		providers: providers,
		handler:   handler,
	}
}
