package renderer

import (
	"context"
	"io"

	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/server/views"
	"github.com/nikimonax/laretz/internal/server/views/pages"
)

func renderHome(ctx context.Context, w io.Writer, data pages.HomeData) error {
	return pages.Home(data).Render(ctx, w)
}

func provideHomeData(
	portfolioList *usecase.ListPortfoliosUseCase,
	providersList *usecase.ListExternalAssetProvidersUseCase,
) func(ctx context.Context) pages.HomeData {
	return func(ctx context.Context) pages.HomeData {
		var data pages.HomeData

		portfolios, err := portfolioList.Execute(ctx)

		if err != nil {
			data.Error = views.Error{Message: "internal error"}
		} else {
			data.Portfolios = portfolios
		}

		data.Providers = providersList.Execute(ctx)

		return data
	}
}

func Home(
	portfolioList *usecase.ListPortfoliosUseCase,
	providersList *usecase.ListExternalAssetProvidersUseCase,
) *ResponseRenderer[pages.HomeData] {
	return NewResponseRenderer(
		RendererFunc[pages.HomeData](renderHome),
		ProviderFunc[pages.HomeData](provideHomeData(portfolioList, providersList)),
	)
}

func HomeWithError(message string, dialogOpt ...string) Mutate[pages.HomeData] {
	dialog := ""

	if len(dialogOpt) > 0 {
		dialog = dialogOpt[0]
	}

	return func(data *pages.HomeData) {
		data.Error = views.Error{Message: message, Dialog: dialog}
	}
}
