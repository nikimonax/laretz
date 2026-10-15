package renderer

import (
	"context"
	"io"

	"github.com/nikimonax/laretz/internal/server/views/pages"
)

const messageSomethingWentWrong = "Something went wrong."

func renderError(ctx context.Context, w io.Writer, data pages.ErrorData) error {
	return pages.ErrorPage(data).Render(ctx, w)
}

func provideErrorData(title string, detail string) func(ctx context.Context) (pages.ErrorData, error) {
	data := pages.ErrorData{Title: title, Detail: detail}
	return func(_ context.Context) (pages.ErrorData, error) { return data, nil }
}

func NotFound() *ResponseRenderer[pages.ErrorData] {
	return NewResponseRenderer(
		RendererFunc[pages.ErrorData](renderError),
		ProviderFunc[pages.ErrorData](provideErrorData("Page not found", "The page you requested does not exist.")),
	)
}

func MethodNotAllowed() *ResponseRenderer[pages.ErrorData] {
	return NewResponseRenderer(
		RendererFunc[pages.ErrorData](renderError),
		ProviderFunc[pages.ErrorData](provideErrorData("Method not allowed", "This address does not accept this request method.")),
	)
}
