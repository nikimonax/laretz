package renderer

import (
	"context"
	"errors"
	"io"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/server/views"
	"github.com/nikimonax/laretz/internal/server/views/pages"
)

func renderAssetSearch(ctx context.Context, w io.Writer, data pages.AssetSearchData) error {
	return pages.AssetSearch(data).Render(ctx, w)
}

func provideAssetSearchData(
	search *usecase.SearchExternalAssetsUseCase,
	assetGet *usecase.GetAssetUseCase,
	externalAssetsCount *usecase.CountExternalAssetsUseCase,
	query string,
	relinkID domain.AssetID,
) func(ctx context.Context) pages.AssetSearchData {
	return func(ctx context.Context) pages.AssetSearchData {
		var data pages.AssetSearchData

		data.Query = query

		count, err := externalAssetsCount.Execute(ctx)

		if err != nil {
			data.Error = views.Error{Message: messageInternalError}
			return data
		}

		data.CatalogEmpty = count == 0

		if relinkID != 0 {
			asset, err := assetGet.Execute(ctx, relinkID)

			if errors.Is(err, app.ErrEntityNotFound) {
				data.Error = views.Error{Message: "Asset not found."}
				return data
			}

			if err != nil {
				data.Error = views.Error{Message: messageInternalError}
				return data
			}

			data.Relink = &pages.AssetRelink{
				AssetID:     relinkID,
				AssetSymbol: asset.Symbol,
				Source:      asset.External.SourceID,
			}
		}

		if query == "" {
			return data
		}

		results, err := search.Execute(ctx, usecase.SearchExternalAssets{Query: query})

		if err != nil {
			data.Error = views.Error{Message: messageInternalError}
			return data
		}

		data.Results = results

		return data
	}
}

func AssetSearch(
	search *usecase.SearchExternalAssetsUseCase,
	assetGet *usecase.GetAssetUseCase,
	externalAssetsCount *usecase.CountExternalAssetsUseCase,
	query string,
	relinkID domain.AssetID,
) *ResponseRenderer[pages.AssetSearchData] {
	return NewResponseRenderer(
		RendererFunc[pages.AssetSearchData](renderAssetSearch),
		ProviderFunc[pages.AssetSearchData](provideAssetSearchData(search, assetGet, externalAssetsCount, query, relinkID)),
	)
}

func AssetSearchWithError(message string, dialogOpt ...string) Mutate[pages.AssetSearchData] {
	dialog := ""

	if len(dialogOpt) > 0 {
		dialog = dialogOpt[0]
	}

	return func(data *pages.AssetSearchData) {
		data.Error = views.Error{Message: message, Dialog: dialog}
	}
}
