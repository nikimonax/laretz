package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/testing/mocks"
)

var errRepo = errors.New("repo failure")

func testAsset(id domain.AssetID, symbol string, externalID domain.ExternalAssetID) domain.Asset {
	return domain.Asset{
		ID:     id,
		Symbol: symbol,
		Name:   symbol + " name",
		External: domain.ExternalAsset{
			ID:         externalID,
			SourceID:   "coingecko",
			ExternalID: symbol,
		},
	}
}

func testOperation(id int64, portfolioID domain.PortfolioID, assetID domain.AssetID, opType domain.OperationType, at time.Time, quantity string) domain.Operation {
	return domain.Operation{
		ID:          domain.OperationID(id),
		PortfolioID: portfolioID,
		AssetID:     assetID,
		Type:        opType,
		ExecutedAt:  at,
		Quantity:    decimal.RequireFromString(quantity),
	}
}

func TestCreateOperationUseCase(t *testing.T) {
	ctx := context.Background()

	t.Run("passes the command to the repository", func(t *testing.T) {
		operations := &mocks.MoqOperationRepository{}

		var created domain.Operation

		operations.CreateFunc = func(_ context.Context, operation domain.Operation) error {
			created = operation
			return nil
		}

		executedAt := time.Now().UTC()

		err := usecase.NewCreateOperationUseCase(operations).Execute(ctx, usecase.CreateOperation{
			PortfolioID: 1,
			AssetID:     2,
			Type:        domain.OperationBuy,
			ExecutedAt:  executedAt,
			Quantity:    decimal.RequireFromString("1.5"),
			TotalUSD:    decimal.RequireFromString("100"),
			Note:        "note",
		})

		require.NoError(t, err)
		assert.Equal(t, domain.PortfolioID(1), created.PortfolioID)
		assert.Equal(t, domain.AssetID(2), created.AssetID)
		assert.Equal(t, domain.OperationBuy, created.Type)
		assert.Equal(t, executedAt, created.ExecutedAt)
		assert.True(t, decimal.RequireFromString("1.5").Equal(created.Quantity))
		assert.True(t, decimal.RequireFromString("100").Equal(created.TotalUSD))
		assert.Equal(t, "note", created.Note)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		operations := &mocks.MoqOperationRepository{}
		operations.CreateFunc = func(context.Context, domain.Operation) error { return errRepo }

		err := usecase.NewCreateOperationUseCase(operations).Execute(ctx, usecase.CreateOperation{})

		require.ErrorIs(t, err, errRepo)
	})
}

func TestUpdateOperationUseCase(t *testing.T) {
	ctx := context.Background()

	t.Run("keeps the operation id", func(t *testing.T) {
		operations := &mocks.MoqOperationRepository{}

		var updated domain.Operation

		operations.UpdateFunc = func(_ context.Context, operation domain.Operation) error {
			updated = operation
			return nil
		}

		err := usecase.NewUpdateOperationUseCase(operations).Execute(ctx, usecase.UpdateOperation{
			ID:              7,
			CreateOperation: usecase.CreateOperation{Type: domain.OperationSell},
		})

		require.NoError(t, err)
		assert.Equal(t, domain.OperationID(7), updated.ID)
		assert.Equal(t, domain.OperationSell, updated.Type)
	})
}

func TestDeleteOperationUseCase(t *testing.T) {
	operations := &mocks.MoqOperationRepository{}

	var deleted domain.OperationID

	operations.DeleteFunc = func(_ context.Context, id domain.OperationID) error {
		deleted = id
		return nil
	}

	err := usecase.NewDeleteOperationUseCase(operations).Execute(context.Background(), usecase.DeleteOperation{ID: 42})

	require.NoError(t, err)
	assert.Equal(t, domain.OperationID(42), deleted)
}

func TestGetOperationUseCase(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the operation", func(t *testing.T) {
		operations := &mocks.MoqOperationRepository{}
		want := testOperation(1, 1, 1, domain.OperationBuy, time.Now(), "1")

		operations.GetFunc = func(context.Context, domain.OperationID) (domain.Operation, error) {
			return want, nil
		}

		got, err := usecase.NewGetOperationUseCase(operations).Execute(ctx, 1)

		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("propagates ErrEntityNotFound", func(t *testing.T) {
		operations := &mocks.MoqOperationRepository{}
		operations.GetFunc = func(context.Context, domain.OperationID) (domain.Operation, error) {
			return domain.Operation{}, app.ErrEntityNotFound
		}

		_, err := usecase.NewGetOperationUseCase(operations).Execute(ctx, 999)

		require.ErrorIs(t, err, app.ErrEntityNotFound)
	})
}

func TestImportAssetUseCase(t *testing.T) {
	assets := &mocks.MoqAssetRepository{}

	var (
		externalID domain.ExternalAssetID
		symbol     string
		name       string
	)

	assets.CreateFunc = func(_ context.Context, e domain.ExternalAssetID, s string, n string) (domain.AssetID, error) {
		externalID, symbol, name = e, s, n
		return 5, nil
	}

	err := usecase.NewImportAssetUseCase(assets).Execute(context.Background(), usecase.ImportAsset{
		ExternalAssetID: 77,
		Symbol:          "BTC",
		Name:            "Bitcoin",
	})

	require.NoError(t, err)
	assert.Equal(t, domain.ExternalAssetID(77), externalID)
	assert.Equal(t, "BTC", symbol)
	assert.Equal(t, "Bitcoin", name)
}

func TestUpdateAssetUseCase(t *testing.T) {
	ctx := context.Background()

	newID := domain.ExternalAssetID(99)

	t.Run("rename and relink together", func(t *testing.T) {
		assets := &mocks.MoqAssetRepository{}

		var (
			renamedID domain.AssetID
			symbol    string
			name      string
			relinkID  domain.AssetID
			relinkTo  domain.ExternalAssetID
		)

		assets.RenameFunc = func(_ context.Context, id domain.AssetID, s string, n string) error {
			renamedID, symbol, name = id, s, n
			return nil
		}

		assets.RelinkFunc = func(_ context.Context, id domain.AssetID, e domain.ExternalAssetID) error {
			relinkID, relinkTo = id, e
			return nil
		}

		err := usecase.NewUpdateAssetUseCase(assets).Execute(ctx, usecase.UpdateAsset{
			ID:              3,
			Symbol:          &[]string{"BTC2"}[0],
			Name:            &[]string{"Bitcoin II"}[0],
			ExternalAssetID: &newID,
		})

		require.NoError(t, err)
		assert.Equal(t, domain.AssetID(3), renamedID)
		assert.Equal(t, "BTC2", symbol)
		assert.Equal(t, "Bitcoin II", name)
		assert.Equal(t, domain.AssetID(3), relinkID)
		assert.Equal(t, newID, relinkTo)
	})

	t.Run("rename without name is rejected", func(t *testing.T) {
		assets := &mocks.MoqAssetRepository{}
		assets.RenameFunc = func(context.Context, domain.AssetID, string, string) error { return nil }

		symbol := "BTC2"

		err := usecase.NewUpdateAssetUseCase(assets).Execute(ctx, usecase.UpdateAsset{ID: 3, Symbol: &symbol})

		require.Error(t, err)
		assert.Empty(t, assets.RenameCalls(), "rename must not be called")
	})

	t.Run("relink only", func(t *testing.T) {
		assets := &mocks.MoqAssetRepository{}

		var relinkTo domain.ExternalAssetID

		assets.RelinkFunc = func(_ context.Context, _ domain.AssetID, e domain.ExternalAssetID) error {
			relinkTo = e
			return nil
		}

		err := usecase.NewUpdateAssetUseCase(assets).Execute(ctx, usecase.UpdateAsset{ID: 3, ExternalAssetID: &newID})

		require.NoError(t, err)
		assert.Equal(t, newID, relinkTo)
	})
}

func TestDeleteAssetUseCase(t *testing.T) {
	ctx := context.Background()

	t.Run("delegates to the repository", func(t *testing.T) {
		assets := &mocks.MoqAssetRepository{}

		var deleted domain.AssetID

		assets.DeleteFunc = func(_ context.Context, id domain.AssetID) error {
			deleted = id
			return nil
		}

		err := usecase.NewDeleteAssetUseCase(assets).Execute(ctx, usecase.DeleteAsset{ID: 4})

		require.NoError(t, err)
		assert.Equal(t, domain.AssetID(4), deleted)
	})

	t.Run("blocks deletion of an asset with operations", func(t *testing.T) {
		assets := &mocks.MoqAssetRepository{}
		assets.DeleteFunc = func(context.Context, domain.AssetID) error { return app.ErrAssetInUse }

		err := usecase.NewDeleteAssetUseCase(assets).Execute(ctx, usecase.DeleteAsset{ID: 4})

		require.ErrorIs(t, err, app.ErrAssetInUse)
	})
}

func TestSearchExternalAssetsUseCase(t *testing.T) {
	assets := &mocks.MoqExternalAssetRepository{}

	var query string

	want := []domain.ExternalAsset{{ID: 1, SourceID: "coingecko", Symbol: "BTC", Name: "Bitcoin"}}

	assets.SearchFunc = func(_ context.Context, q string) ([]domain.ExternalAsset, error) {
		query = q
		return want, nil
	}

	got, err := usecase.NewSearchExternalAssetsUseCase(assets).Execute(context.Background(), usecase.SearchExternalAssets{Query: "bit"})

	require.NoError(t, err)
	assert.Equal(t, "bit", query)
	assert.Equal(t, want, got)
}

func TestSyncExternalAssetsUseCase(t *testing.T) {
	ctx := context.Background()

	catalog := []domain.ExternalAsset{
		{ID: 1, SourceID: "coingecko", ExternalID: "bitcoin", Symbol: "BTC", Name: "Bitcoin"},
	}

	t.Run("syncs the provider catalog", func(t *testing.T) {
		providers := &mocks.MoqExternalAssetProviderRegistry{}
		provider := &mocks.MoqExternalAssetProvider{}
		assets := &mocks.MoqExternalAssetRepository{}

		providers.FindFunc = func(domain.SourceID) (interfaces.ExternalAssetProvider, error) {
			return provider, nil
		}
		provider.ProvideFunc = func(context.Context) ([]domain.ExternalAsset, error) { return catalog, nil }

		var synced []domain.ExternalAsset

		assets.SyncFunc = func(_ context.Context, got []domain.ExternalAsset) error {
			synced = got
			return nil
		}

		err := usecase.NewSyncExternalAssetsUseCase(providers, assets).Execute(ctx, usecase.SyncExternalAssets{SourceID: "coingecko"})

		require.NoError(t, err)
		assert.Equal(t, catalog, synced)
	})

	t.Run("unknown source fails before fetching", func(t *testing.T) {
		providers := &mocks.MoqExternalAssetProviderRegistry{}
		assets := &mocks.MoqExternalAssetRepository{}

		providers.FindFunc = func(domain.SourceID) (interfaces.ExternalAssetProvider, error) {
			return nil, app.ErrEntityNotFound
		}

		err := usecase.NewSyncExternalAssetsUseCase(providers, assets).Execute(ctx, usecase.SyncExternalAssets{SourceID: "nope"})

		require.ErrorIs(t, err, app.ErrEntityNotFound)
		assert.Empty(t, assets.SyncCalls())
	})

	t.Run("provider failure propagates", func(t *testing.T) {
		providers := &mocks.MoqExternalAssetProviderRegistry{}
		provider := &mocks.MoqExternalAssetProvider{}
		assets := &mocks.MoqExternalAssetRepository{}

		providers.FindFunc = func(domain.SourceID) (interfaces.ExternalAssetProvider, error) { return provider, nil }
		provider.ProvideFunc = func(context.Context) ([]domain.ExternalAsset, error) { return nil, errRepo }

		err := usecase.NewSyncExternalAssetsUseCase(providers, assets).Execute(ctx, usecase.SyncExternalAssets{SourceID: "coingecko"})

		require.ErrorIs(t, err, errRepo)
		assert.Empty(t, assets.SyncCalls())
	})
}

func TestCountExternalAssetsUseCase(t *testing.T) {
	assets := &mocks.MoqExternalAssetRepository{}
	assets.CountFunc = func(context.Context) (int64, error) { return 7, nil }

	count, err := usecase.NewCountExternalAssetsUseCase(assets).Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, int64(7), count)
}

func TestListExternalAssetProvidersUseCase(t *testing.T) {
	providers := &mocks.MoqExternalAssetProviderRegistry{}
	providers.ListFunc = func() []domain.SourceID { return []domain.SourceID{"coingecko"} }

	assert.Equal(t, []domain.SourceID{"coingecko"}, usecase.NewListExternalAssetProvidersUseCase(providers).Execute(context.Background()))
}
