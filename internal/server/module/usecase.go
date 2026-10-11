package module

import (
	"go.uber.org/fx"

	"github.com/nikimonax/laretz/internal/app/usecase"
)

// depends on repository module
func UseCaseModule() fx.Option {
	return fx.Module(
		"usecase",
		fx.Provide(
			usecase.NewCreatePortfolioUseCase,
			usecase.NewListPortfoliosUseCase,
			usecase.NewUpdatePortfolioUseCase,
			usecase.NewDeletePortfolioUseCase,
			usecase.NewSyncExternalAssetsUseCase,
			usecase.NewListExternalAssetProvidersUseCase,
			usecase.NewSearchExternalAssetsUseCase,
			usecase.NewCountExternalAssetsUseCase,
			usecase.NewListAssetsUseCase,
			usecase.NewGetAssetUseCase,
			usecase.NewImportAssetUseCase,
			usecase.NewUpdateAssetUseCase,
			usecase.NewDeleteAssetUseCase,
		),
	)
}
