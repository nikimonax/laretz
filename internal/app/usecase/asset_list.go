package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type ListAssetsUseCase struct {
	repo interfaces.AssetRepository
}

func (usecase *ListAssetsUseCase) Execute(ctx context.Context) ([]domain.Asset, error) {
	return usecase.repo.List(ctx)
}

func NewListAssetsUseCase(repo interfaces.AssetRepository) *ListAssetsUseCase {
	return &ListAssetsUseCase{repo: repo}
}
