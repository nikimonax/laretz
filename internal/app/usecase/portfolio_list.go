package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type ListPortfoliosUseCase struct {
	repo interfaces.PortfolioRepository
}

func NewListPortfoliosUseCase(repo interfaces.PortfolioRepository) *ListPortfoliosUseCase {
	return &ListPortfoliosUseCase{repo: repo}
}

func (usecase *ListPortfoliosUseCase) Execute(ctx context.Context) ([]domain.Portfolio, error) {
	return usecase.repo.List(ctx)
}
