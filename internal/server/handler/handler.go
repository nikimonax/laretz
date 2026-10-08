package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/lib/spoof"
	"github.com/nikimonax/laretz/internal/server/middleware"
	"github.com/nikimonax/laretz/internal/server/renderer"
	"github.com/nikimonax/laretz/internal/server/static"
)

func New(
	logger logr.Logger,
	portfolioList *usecase.ListPortfoliosUseCase,
	providersList *usecase.ListExternalAssetProvidersUseCase,
	portfolioCreate *usecase.CreatePortfolioUseCase,
	portfolioUpdate *usecase.UpdatePortfolioUseCase,
	portfolioDelele *usecase.DeletePortfolioUseCase,
	externalAssetsSync *usecase.SyncExternalAssetsUseCase,
) chi.Router {
	router := chi.NewRouter()

	router.Use(
		middleware.OriginCheck,
		spoof.Middleware,
		middleware.NewLoggingMiddleware(logger),
	)

	homeRenderer := renderer.Home(portfolioList, providersList)

	notFoundHandler := NewPageHandler(logger, renderer.NotFound())
	methodNotAllowedHandler := NewPageHandler(logger, renderer.MethodNotAllowed())

	homeHandler := NewPageHandler(logger, homeRenderer)
	portfolioCreateHandler := NewCreatePortfolioHandler(logger, homeRenderer, portfolioCreate)
	portfolioUpdateHandler := NewUpdatePortfolioHandler(logger, homeRenderer, portfolioUpdate)
	portfolioDeleleHandler := NewDeletePortfolioHandler(logger, homeRenderer, portfolioDelele)
	syncExternalAssetsHandler := NewSyncExternalAssetsHandler(logger, homeRenderer, externalAssetsSync)

	static.Mount(router, notFoundHandler)
	router.Get("/", homeHandler.ServeHTTP)
	router.Post("/portfolio", portfolioCreateHandler.ServeHTTP)
	router.Post("/portfolio/{id}", portfolioUpdateHandler.ServeHTTP)
	router.Delete("/portfolio/{id}", portfolioDeleleHandler.ServeHTTP)
	router.Post("/assets/sync", syncExternalAssetsHandler.ServeHTTP)

	router.NotFound(notFoundHandler.ServeHTTP)
	router.MethodNotAllowed(methodNotAllowedHandler.ServeHTTP)

	return router
}
