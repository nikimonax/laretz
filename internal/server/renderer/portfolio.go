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

func renderPortfolio(ctx context.Context, w io.Writer, data pages.PortfolioData) error {
	return pages.Portfolio(data).Render(ctx, w)
}

func providePortfolioData(
	portfolioGet *usecase.GetPortfolioUseCase,
	assetsList *usecase.ListAssetsUseCase,
	portfolioID domain.PortfolioID,
	opsPage int,
) func(ctx context.Context) (pages.PortfolioData, error) {
	return func(ctx context.Context) (pages.PortfolioData, error) {
		var data pages.PortfolioData

		detail, err := portfolioGet.Execute(ctx, usecase.GetPortfolio{
			ID:      portfolioID,
			OpsPage: opsPage,
		})

		if err != nil {
			message := messageSomethingWentWrong

			if errors.Is(err, app.ErrEntityNotFound) {
				message = "Portfolio not found."
			}

			data.Error = views.Error{Message: message}
			return data, fmt.Errorf("get portfolio %d: %w", portfolioID, err)
		}

		data.Portfolio = detail.Portfolio
		data.Total = detail.Total
		data.Positions = detail.Positions
		data.Operations = detail.Operations.Items
		data.OpsPage = detail.Operations.Page
		data.OpsPages = detail.Operations.Pages

		assets, err := assetsList.Execute(ctx)

		if err != nil {
			data.Error = views.Error{Message: messageSomethingWentWrong}
			return data, fmt.Errorf("list assets: %w", err)
		}

		data.Assets = assets

		return data, nil
	}
}

func Portfolio(
	portfolioGet *usecase.GetPortfolioUseCase,
	assetsList *usecase.ListAssetsUseCase,
	portfolioID domain.PortfolioID,
	opsPage int,
) *ResponseRenderer[pages.PortfolioData] {
	return NewResponseRenderer(
		RendererFunc[pages.PortfolioData](renderPortfolio),
		ProviderFunc[pages.PortfolioData](providePortfolioData(portfolioGet, assetsList, portfolioID, opsPage)),
	)
}

func PortfolioWithError(message string, dialogOpt ...string) Mutate[pages.PortfolioData] {
	dialog := ""

	if len(dialogOpt) > 0 {
		dialog = dialogOpt[0]
	}

	return func(data *pages.PortfolioData) {
		data.Error = views.Error{Message: message, Dialog: dialog}
	}
}
