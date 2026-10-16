package renderer

import (
	"context"
	"fmt"
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
) func(ctx context.Context) (pages.HomeData, error) {
	return func(ctx context.Context) (pages.HomeData, error) {
		var data pages.HomeData

		summaries, err := portfolioList.Execute(ctx)

		if err != nil {
			data.Error = views.Error{Message: messageSomethingWentWrong}
			return data, fmt.Errorf("list portfolios: %w", err)
		}

		data.Portfolios = summaries

		return data, nil
	}
}

func Home(
	portfolioList *usecase.ListPortfoliosUseCase,
) *ResponseRenderer[pages.HomeData] {
	return NewResponseRenderer(
		RendererFunc[pages.HomeData](renderHome),
		ProviderFunc[pages.HomeData](provideHomeData(portfolioList)),
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
