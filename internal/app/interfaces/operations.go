package interfaces

import (
	"context"

	"github.com/nikimonax/laretz/internal/domain"
)

type OperationRepository interface {
	Get(ctx context.Context, id domain.OperationID) (domain.Operation, error)
	List(ctx context.Context, portfolioID domain.PortfolioID) ([]domain.Operation, error)
	ListPaged(ctx context.Context, portfolioID domain.PortfolioID, limit, offset int) ([]domain.Operation, error)
	Count(ctx context.Context, portfolioID domain.PortfolioID) (int64, error)
	Create(ctx context.Context, operation domain.Operation) error
	Update(ctx context.Context, operation domain.Operation) error
	Delete(ctx context.Context, id domain.OperationID) error
}
