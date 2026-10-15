package renderer

import (
	"context"
	"errors"
	"fmt"
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
) func(ctx context.Context) (pages.AssetSearchData, error) {
	return func(ctx context.Context) (pages.AssetSearchData, error) {
		var data pages.AssetSearchData

		data.Query = query

		count, err := externalAssetsCount.Execute(ctx)

		if err != nil {
			data.Error = views.Error{Message: messageSomethingWentWrong}
			return data, fmt.Errorf("count external assets: %w", err)
		}

		data.CatalogEmpty = count == 0

		if relinkID != 0 {
			asset, err := assetGet.Execute(ctx, relinkID)

			if err != nil {
				message := messageSomethingWentWrong

				if errors.Is(err, app.ErrEntityNotFound) {
					message = "Asset not found."
				}

				data.Error = views.Error{Message: message}
				return data, fmt.Errorf("get asset %d: %w", relinkID, err)
			}

			data.Relink = &pages.AssetRelink{
				AssetID:     relinkID,
				AssetSymbol: asset.Symbol,
				Source:      asset.External.SourceID,
			}
		}

		if query == "" {
			return data, nil
		}

		results, err := search.Execute(ctx, usecase.SearchExternalAssets{Query: query})

		if err != nil {
			data.Error = views.Error{Message: messageSomethingWentWrong}
			return data, fmt.Errorf("search external assets: %w", err)
		}

		data.Results = results

		return data, nil
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
