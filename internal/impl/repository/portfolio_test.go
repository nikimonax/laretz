package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/impl/repository"
	"github.com/nikimonax/laretz/internal/testing/db"
)

func TestPortfolioRepository(t *testing.T) {
	ctx := context.Background()
	portfolios := repository.NewSqlitePortfolioRepository(db.New(t))

	t.Run("create get rename list delete round trip", func(t *testing.T) {
		created, err := portfolios.Create(ctx, "main")
		require.NoError(t, err)

		assert.Positive(t, created.ID)
		assert.Equal(t, "main", created.Name)
		assert.False(t, created.CreatedAt.IsZero())
		assert.False(t, created.UpdatedAt.IsZero())

		got, err := portfolios.Get(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, "main", got.Name)

		require.NoError(t, portfolios.Rename(ctx, created.ID, "renamed"))
		got, err = portfolios.Get(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "renamed", got.Name)
		assert.False(t, got.UpdatedAt.Before(got.CreatedAt))

		second, err := portfolios.Create(ctx, "second")
		require.NoError(t, err)

		list, err := portfolios.List(ctx)
		require.NoError(t, err)
		require.Len(t, list, 2)
		assert.Equal(t, []string{"renamed", "second"}, []string{list[0].Name, list[1].Name})

		require.NoError(t, portfolios.Delete(ctx, created.ID))

		list, err = portfolios.List(ctx)
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, second.ID, list[0].ID)
	})

	t.Run("missing entities map to ErrEntityNotFound", func(t *testing.T) {
		_, err := portfolios.Get(ctx, 424242)
		require.ErrorIs(t, err, app.ErrEntityNotFound)

		require.ErrorIs(t, portfolios.Rename(ctx, 424242, "x"), app.ErrEntityNotFound)
		require.ErrorIs(t, portfolios.Delete(ctx, 424242), app.ErrEntityNotFound)
	})

	t.Run("list on an empty database returns an empty result", func(t *testing.T) {
		portfolios := repository.NewSqlitePortfolioRepository(db.New(t))

		list, err := portfolios.List(ctx)
		require.NoError(t, err)
		assert.Empty(t, list)
	})
}
