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
	providersList *usecase.ListExternalAssetProvidersUseCase,
	portfolioList *usecase.ListPortfoliosUseCase,
	portfolioGet *usecase.GetPortfolioUseCase,
	portfolioCreate *usecase.CreatePortfolioUseCase,
	portfolioUpdate *usecase.UpdatePortfolioUseCase,
	portfolioDelele *usecase.DeletePortfolioUseCase,
	operationCreate *usecase.CreateOperationUseCase,
	operationUpdate *usecase.UpdateOperationUseCase,
	operationDelete *usecase.DeleteOperationUseCase,
	externalAssetsSync *usecase.SyncExternalAssetsUseCase,
	externalAssetsCount *usecase.CountExternalAssetsUseCase,
	externalAssetsSearch *usecase.SearchExternalAssetsUseCase,
	assetsList *usecase.ListAssetsUseCase,
	assetGet *usecase.GetAssetUseCase,
	assetImport *usecase.ImportAssetUseCase,
	assetUpdate *usecase.UpdateAssetUseCase,
	assetDelete *usecase.DeleteAssetUseCase,
) chi.Router {
	router := chi.NewRouter()

	router.Use(
		middleware.OriginCheck,
		spoof.Middleware,
		middleware.NewLoggingMiddleware(logger),
	)

	homeRenderer := renderer.Home(portfolioList)
	assetsRenderer := renderer.Assets(assetsList, externalAssetsCount, providersList)

	notFoundHandler := NewPageHandler(logger, renderer.NotFound())
	methodNotAllowedHandler := NewPageHandler(logger, renderer.MethodNotAllowed())

	homeHandler := NewPageHandler(logger, homeRenderer)
	assetsHandler := NewPageHandler(logger, assetsRenderer)
	assetSearchHandler := NewAssetSearchHandler(logger, externalAssetsSearch, assetGet, externalAssetsCount)
	portfolioGetHandler := NewPortfolioGetHandler(logger, portfolioGet, assetsList)
	portfolioCreateHandler := NewCreatePortfolioHandler(logger, homeRenderer, portfolioCreate)
	portfolioUpdateHandler := NewUpdatePortfolioHandler(logger, homeRenderer, portfolioUpdate)
	portfolioDeleleHandler := NewDeletePortfolioHandler(logger, homeRenderer, portfolioDelele)
	createOperationHandler := NewCreateOperationHandler(logger, portfolioGet, assetsList, operationCreate)
	updateOperationHandler := NewUpdateOperationHandler(logger, portfolioGet, assetsList, operationUpdate)
	deleteOperationHandler := NewDeleteOperationHandler(logger, portfolioGet, assetsList, operationDelete)
	syncExternalAssetsHandler := NewSyncExternalAssetsHandler(logger, assetsRenderer, externalAssetsSync)
	importAssetHandler := NewImportAssetHandler(logger, externalAssetsSearch, assetGet, externalAssetsCount, assetImport)
	updateAssetHandler := NewUpdateAssetHandler(logger, assetsRenderer, externalAssetsSearch, assetGet, externalAssetsCount, assetUpdate)
	deleteAssetHandler := NewDeleteAssetHandler(logger, assetsRenderer, assetDelete)

	static.Mount(router, notFoundHandler)
	router.Get("/", homeHandler.ServeHTTP)
	router.Get("/portfolio/{id}", portfolioGetHandler.ServeHTTP)
	router.Post("/portfolio", portfolioCreateHandler.ServeHTTP)
	router.Post("/portfolio/{id}", portfolioUpdateHandler.ServeHTTP)
	router.Delete("/portfolio/{id}", portfolioDeleleHandler.ServeHTTP)
	router.Post("/portfolio/{id}/operation", createOperationHandler.ServeHTTP)
	router.Post("/portfolio/{id}/operation/{operationID}", updateOperationHandler.ServeHTTP)
	router.Delete("/portfolio/{id}/operation/{operationID}", deleteOperationHandler.ServeHTTP)
	router.Post("/assets/sync", syncExternalAssetsHandler.ServeHTTP)
	router.Get("/assets", assetsHandler.ServeHTTP)
	router.Get("/assets/search", assetSearchHandler.ServeHTTP)
	router.Post("/assets", importAssetHandler.ServeHTTP)
	router.Post("/assets/{id}", updateAssetHandler.ServeHTTP)
	router.Delete("/assets/{id}", deleteAssetHandler.ServeHTTP)

	router.NotFound(notFoundHandler.ServeHTTP)
	router.MethodNotAllowed(methodNotAllowedHandler.ServeHTTP)

	return router
}
