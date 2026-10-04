package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type UpdatePortfolio struct {
	ID   domain.PortfolioID
	Name string
}

type UpdatePortfolioUseCase struct {
	repo interfaces.PortfolioRepository
}

func NewUpdatePortfolioUseCase(repo interfaces.PortfolioRepository) *UpdatePortfolioUseCase {
	return &UpdatePortfolioUseCase{repo: repo}
}

func (usecase *UpdatePortfolioUseCase) Execute(ctx context.Context, cmd UpdatePortfolio) error {
	return usecase.repo.Rename(ctx, cmd.ID, cmd.Name)
}
