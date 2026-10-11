package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/impl/database/generated"
)

type SqlitePortfolioRepository struct {
	queries *generated.Queries
}

// Get implements [interfaces.PortfolioRepository].
func (repo *SqlitePortfolioRepository) Get(ctx context.Context, id int64) (domain.Portfolio, error) {
	row, err := repo.queries.GetPortfolio(ctx, id)

	if errors.Is(err, sql.ErrNoRows) {
		return domain.Portfolio{}, app.ErrEntityNotFound
	}

	if err != nil {
		return domain.Portfolio{}, fmt.Errorf("get portfolio %d: %w", id, err)
	}

	return toDomainPortfolio(row), nil
}

// Create implements [interfaces.PortfolioRepository].
func (repo *SqlitePortfolioRepository) Create(ctx context.Context, name string) (domain.Portfolio, error) {
	created, err := repo.queries.CreatePortfolio(ctx, name)

	if err != nil {
		return domain.Portfolio{}, fmt.Errorf("create portfolio: %w", err)
	}

	return toDomainPortfolio(created), nil
}

// List implements [interfaces.PortfolioRepository].
func (repo *SqlitePortfolioRepository) List(ctx context.Context) ([]domain.Portfolio, error) {
	rows, err := repo.queries.ListPortfolios(ctx)

	if err != nil {
		return nil, fmt.Errorf("list portfolios: %w", err)
	}

	portfolios := make([]domain.Portfolio, 0, len(rows))

	for _, row := range rows {
		portfolios = append(portfolios, toDomainPortfolio(row))
	}

	return portfolios, nil
}

// Delete implements [interfaces.PortfolioRepository].
func (repo *SqlitePortfolioRepository) Delete(ctx context.Context, id int64) error {
	rows, err := repo.queries.DeletePortfolio(ctx, id)

	if err != nil {
		return fmt.Errorf("delete portfolio %d: %w", id, err)
	}

	if rows == 0 {
		return app.ErrEntityNotFound
	}

	return nil
}

// Rename implements [interfaces.PortfolioRepository].
func (repo *SqlitePortfolioRepository) Rename(ctx context.Context, id int64, name string) error {
	params := generated.RenamePortfolioParams{
		ID:   id,
		Name: name,
	}

	rows, err := repo.queries.RenamePortfolio(ctx, params)

	if err != nil {
		return fmt.Errorf("rename portfolio %d: %w", id, err)
	}

	if rows == 0 {
		return app.ErrEntityNotFound
	}

	return nil
}

func NewSqlitePortfolioRepository(db *sql.DB) interfaces.PortfolioRepository {
	return &SqlitePortfolioRepository{queries: generated.New(db)}
}

func toDomainPortfolio(row generated.Portfolio) domain.Portfolio {
	return domain.Portfolio{
		ID:        row.ID,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
