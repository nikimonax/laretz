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
			usecase.NewGetPortfolioUseCase,
			usecase.NewGetPortfolioChartUseCase,
			usecase.NewUpdatePortfolioUseCase,
			usecase.NewDeletePortfolioUseCase,
			usecase.NewCreateOperationUseCase,
			usecase.NewUpdateOperationUseCase,
			usecase.NewDeleteOperationUseCase,
			usecase.NewGetOperationUseCase,
			usecase.NewSyncExternalAssetsUseCase,
			usecase.NewListExternalAssetProvidersUseCase,
			usecase.NewSearchExternalAssetsUseCase,
			usecase.NewCountExternalAssetsUseCase,
			usecase.NewListAssetsUseCase,
			usecase.NewGetAssetUseCase,
			usecase.NewImportAssetUseCase,
			usecase.NewUpdateAssetUseCase,
			usecase.NewDeleteAssetUseCase,
			usecase.NewFetchQuotesUseCase,
		),
	)
}
