package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

const (
	sourceCoingecko = "coingecko"
	sourceBinance   = "binance"
)

func seedPortfolio(ctx context.Context, t *testing.T, portfolios interfaces.PortfolioRepository) domain.Portfolio {
	t.Helper()

	created, err := portfolios.Create(ctx, "test")
	require.NoError(t, err)

	return created
}

func seedExternalAsset(
	ctx context.Context,
	t *testing.T,
	externalAssets interfaces.ExternalAssetRepository,
	sourceID domain.SourceID,
	externalID string,
	symbol string,
) domain.ExternalAsset {
	t.Helper()

	require.NoError(t, externalAssets.Sync(ctx, []domain.ExternalAsset{
		{SourceID: sourceID, ExternalID: externalID, Symbol: symbol, Name: symbol + " name"},
	}))

	found, err := externalAssets.Search(ctx, symbol)
	require.NoError(t, err)

	for _, asset := range found {
		if asset.SourceID == sourceID && asset.ExternalID == externalID {
			return asset
		}
	}

	require.FailNow(t, "seeded external asset not found by search")

	return domain.ExternalAsset{}
}

func seedAsset(
	ctx context.Context,
	t *testing.T,
	assets interfaces.AssetRepository,
	externalAsset domain.ExternalAsset,
	symbol string,
) domain.Asset {
	t.Helper()

	id, err := assets.Create(ctx, externalAsset.ID, symbol, symbol+" name")
	require.NoError(t, err)

	asset, err := assets.Get(ctx, id)
	require.NoError(t, err)

	return asset
}
