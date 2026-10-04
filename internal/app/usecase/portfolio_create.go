package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type CreatePorfolio struct {
	Name string
}

type CreatePortfolioUseCase struct {
	repo interfaces.PortfolioRepository
}

func NewCreatePortfolioUseCase(repo interfaces.PortfolioRepository) *CreatePortfolioUseCase {
	return &CreatePortfolioUseCase{repo: repo}
}

func (usecase *CreatePortfolioUseCase) Execute(ctx context.Context, cmd CreatePorfolio) (domain.Portfolio, error) {
	return usecase.repo.Create(ctx, cmd.Name)
}
