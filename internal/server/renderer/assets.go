package renderer

import (
	"context"
	"io"

	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/server/views"
	"github.com/nikimonax/laretz/internal/server/views/pages"
)

func renderAssets(ctx context.Context, w io.Writer, data pages.AssetsData) error {
	return pages.Assets(data).Render(ctx, w)
}

func provideAssetsData(
	assetsList *usecase.ListAssetsUseCase,
	externalAssetsCount *usecase.CountExternalAssetsUseCase,
	providersList *usecase.ListExternalAssetProvidersUseCase,
) func(ctx context.Context) pages.AssetsData {
	return func(ctx context.Context) pages.AssetsData {
		var data pages.AssetsData

		assets, err := assetsList.Execute(ctx)

		if err != nil {
			data.Error = views.Error{Message: messageInternalError}
		} else {
			data.Assets = assets
		}

		count, err := externalAssetsCount.Execute(ctx)

		if err != nil {
			data.Error = views.Error{Message: messageInternalError}
		} else {
			data.CatalogEmpty = count == 0
		}

		data.Providers = providersList.Execute(ctx)

		return data
	}
}

func Assets(
	assetsList *usecase.ListAssetsUseCase,
	externalAssetsCount *usecase.CountExternalAssetsUseCase,
	providersList *usecase.ListExternalAssetProvidersUseCase,
) *ResponseRenderer[pages.AssetsData] {
	return NewResponseRenderer(
		RendererFunc[pages.AssetsData](renderAssets),
		ProviderFunc[pages.AssetsData](provideAssetsData(assetsList, externalAssetsCount, providersList)),
	)
}

func AssetsWithError(message string, dialogOpt ...string) Mutate[pages.AssetsData] {
	dialog := ""

	if len(dialogOpt) > 0 {
		dialog = dialogOpt[0]
	}

	return func(data *pages.AssetsData) {
		data.Error = views.Error{Message: message, Dialog: dialog}
	}
}
