package interfaces

import (
	"context"

	"github.com/nikimonax/laretz/internal/domain"
)

type PortfolioRepository interface {
	Get(ctx context.Context, id int64) (domain.Portfolio, error)
	Create(ctx context.Context, name string) (domain.Portfolio, error)
	List(ctx context.Context) ([]domain.Portfolio, error)
	Rename(ctx context.Context, id int64, name string) error
	Delete(ctx context.Context, id int64) error
}
