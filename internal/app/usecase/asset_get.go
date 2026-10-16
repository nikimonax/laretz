package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type GetAssetUseCase struct {
	repo interfaces.AssetRepository
}

func (usecase *GetAssetUseCase) Execute(ctx context.Context, id domain.AssetID) (domain.Asset, error) {
	return usecase.repo.Get(ctx, id)
}

func NewGetAssetUseCase(repo interfaces.AssetRepository) *GetAssetUseCase {
	return &GetAssetUseCase{repo: repo}
}
