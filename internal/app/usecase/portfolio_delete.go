package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type DeletePortfolio struct {
	ID domain.PortfolioID
}

type DeletePortfolioUseCase struct {
	repo interfaces.PortfolioRepository
}

func NewDeletePortfolioUseCase(repo interfaces.PortfolioRepository) *DeletePortfolioUseCase {
	return &DeletePortfolioUseCase{repo: repo}
}

func (usecase *DeletePortfolioUseCase) Execute(ctx context.Context, cmd DeletePortfolio) error {
	return usecase.repo.Delete(ctx, cmd.ID)
}
