package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type ListExternalAssetProvidersUseCase struct {
	providers interfaces.ExternalAssetProviderRegistry
}

func (usecase *ListExternalAssetProvidersUseCase) Execute(context.Context) []domain.SourceID {
	return usecase.providers.List()
}

func NewListExternalAssetProvidersUseCase(
	providers interfaces.ExternalAssetProviderRegistry,
) *ListExternalAssetProvidersUseCase {
	return &ListExternalAssetProvidersUseCase{providers: providers}
}
