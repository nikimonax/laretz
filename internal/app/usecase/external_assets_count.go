package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
)

type CountExternalAssetsUseCase struct {
	repo interfaces.ExternalAssetRepository
}

func (usecase *CountExternalAssetsUseCase) Execute(ctx context.Context) (int64, error) {
	return usecase.repo.Count(ctx)
}

func NewCountExternalAssetsUseCase(repo interfaces.ExternalAssetRepository) *CountExternalAssetsUseCase {
	return &CountExternalAssetsUseCase{repo: repo}
}
