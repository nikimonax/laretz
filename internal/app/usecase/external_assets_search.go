package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type SearchExternalAssets struct {
	Query string
}

type SearchExternalAssetsUseCase struct {
	repo interfaces.ExternalAssetRepository
}

func (usecase *SearchExternalAssetsUseCase) Execute(ctx context.Context, cmd SearchExternalAssets) ([]domain.ExternalAsset, error) {
	return usecase.repo.Search(ctx, cmd.Query)
}

func NewSearchExternalAssetsUseCase(repo interfaces.ExternalAssetRepository) *SearchExternalAssetsUseCase {
	return &SearchExternalAssetsUseCase{repo: repo}
}
