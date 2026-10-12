package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nikimonax/laretz/internal/domain"
)

func valuedAsset(id domain.AssetID, symbol string, externalID domain.ExternalAssetID, operations []domain.Operation, points domain.PriceHistory) domain.ValuedAsset {
	external := domain.ExternalAsset{ID: externalID}
	if externalID != 0 {
		external.SourceID = "coingecko"
		external.ExternalID = symbol
	}

	return domain.ValuedAsset{
		Asset:    domain.Asset{ID: id, Symbol: symbol, External: external},
		Timeline: domain.NewBalanceTimeline(operations),
		Prices:   points,
	}
}

func TestValuation_At(t *testing.T) {
	at := hour(t, "2026-01-01T10:00:00Z")
	later := at.Add(2 * time.Hour)

	tests := map[string]struct {
		assets        []domain.ValuedAsset
		wantAmount    string
		wantStale     bool
		wantMissing   bool
		wantPositions int
	}{
		"empty portfolio": {
			assets:        nil,
			wantAmount:    "0",
			wantPositions: 0,
		},
		"balance times price": {
			assets: []domain.ValuedAsset{
				valuedAsset(1, "BTC", 10, []domain.Operation{buy(1, at, "2")},
					domain.PriceHistory{
						{At: at, PriceUSD: dec(t, "100"), Granularity: domain.GranularityHour},
						{At: later, PriceUSD: dec(t, "100"), Granularity: domain.GranularityHour},
					}),
			},
			wantAmount: "200", wantPositions: 1,
		},
		"negative balance reduces the value": {
			assets: []domain.ValuedAsset{
				valuedAsset(1, "BTC", 10, []domain.Operation{sell(1, at, "2")},
					domain.PriceHistory{
						{At: at, PriceUSD: dec(t, "100"), Granularity: domain.GranularityHour},
						{At: later, PriceUSD: dec(t, "100"), Granularity: domain.GranularityHour},
					}),
			},
			wantAmount: "-200", wantPositions: 1,
		},
		"carried-forward price is stale but counted": {
			assets: []domain.ValuedAsset{
				valuedAsset(1, "BTC", 10, []domain.Operation{buy(1, at, "1")},
					domain.PriceHistory{{At: at, PriceUSD: dec(t, "100"), Granularity: domain.GranularityHour}}),
			},
			wantAmount: "100", wantStale: true, wantPositions: 1,
		},
		"asset without prices contributes nothing and flags missing": {
			assets: []domain.ValuedAsset{
				valuedAsset(1, "BTC", 10, []domain.Operation{buy(1, at, "1")}, nil),
				valuedAsset(2, "ETH", 20, []domain.Operation{buy(2, at, "1")},
					domain.PriceHistory{{At: at, PriceUSD: dec(t, "10"), Granularity: domain.GranularityHour}}),
			},
			wantAmount: "10", wantStale: true, wantMissing: true, wantPositions: 2,
		},
		"zero balance is not held": {
			assets: []domain.ValuedAsset{
				valuedAsset(1, "BTC", 10, []domain.Operation{buy(1, at, "1"), sell(2, at, "1")}, nil),
			},
			wantAmount:    "0",
			wantPositions: 0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			value, positions := domain.NewValuation(test.assets).At(later)

			assert.True(t, dec(t, test.wantAmount).Equal(value.Amount))
			assert.Equal(t, test.wantStale, value.Stale, "stale")
			assert.Equal(t, test.wantMissing, value.MissingPrices)
			assert.Len(t, positions, test.wantPositions)

			if test.wantMissing {
				var found bool

				for _, position := range positions {
					if position.Asset.Symbol == "BTC" {
						found = true
						assert.False(t, position.Priced)
					}
				}

				assert.True(t, found, "expected an unpriced BTC position")
			}
		})
	}
}

func TestValuation_Series(t *testing.T) {
	at := hour(t, "2026-01-01T10:00:00Z")
	until := at.Add(3 * time.Hour)

	t.Run("forward fill bridges a gap between intervals", func(t *testing.T) {
		valuation := domain.NewValuation([]domain.ValuedAsset{
			valuedAsset(1, "BTC", 10,
				[]domain.Operation{
					buy(1, at, "1"),
					sell(2, at.Add(time.Hour), "1"),
					buy(3, at.Add(2*time.Hour), "1"),
				},
				domain.PriceHistory{{At: at, PriceUSD: dec(t, "100"), Granularity: domain.GranularityHour}}),
		})

		series := valuation.Series(until)

		assert.Len(t, series, 4)
		assert.Equal(t, at, series[0].Hour)
		assert.True(t, series[0].Held)
		assert.True(t, dec(t, "100").Equal(series[0].Amount))
		assert.True(t, series[1].Held)
		assert.True(t, dec(t, "0").Equal(series[1].Amount))
		assert.True(t, series[2].Held)
		assert.True(t, dec(t, "100").Equal(series[2].Amount))
		assert.True(t, series[3].Held)
		assert.True(t, series[3].Stale)
		assert.True(t, dec(t, "100").Equal(series[3].Amount))
	})

	t.Run("no holdings yields nil series", func(t *testing.T) {
		assert.Nil(t, domain.NewValuation(nil).Series(until))
	})

	t.Run("backdated operation extends the series start", func(t *testing.T) {
		valuation := domain.NewValuation([]domain.ValuedAsset{
			valuedAsset(1, "BTC", 10,
				[]domain.Operation{buy(1, at, "1"), buy(2, at.Add(-2*time.Hour), "1")}, nil),
		})

		series := valuation.Series(until)

		assert.Equal(t, at.Add(-2*time.Hour), series[0].Hour)
		for _, point := range series {
			assert.True(t, point.MissingPrices)
		}
	})
}

func TestPriceHistory_PriceAt(t *testing.T) {
	at10 := hour(t, "2026-01-01T10:00:00Z")
	at12 := hour(t, "2026-01-01T12:00:00Z")

	history := domain.PriceHistory{
		{At: at10, PriceUSD: dec(t, "100"), Granularity: domain.GranularityHour},
		{At: at12, PriceUSD: dec(t, "120"), Granularity: domain.GranularityHour},
	}

	tests := map[string]struct {
		hour  time.Time
		price string
		asOf  time.Time
		ok    bool
		stale bool
	}{
		"exact hour":               {hour: at12, price: "120", asOf: at12, ok: true},
		"carried forward":          {hour: at12.Add(time.Hour), price: "120", asOf: at12, ok: true, stale: true},
		"before first point":       {hour: at10.Add(-time.Hour), ok: false},
		"between points":           {hour: at10.Add(time.Hour), price: "100", asOf: at10, ok: true, stale: true},
		"hour itself is not stale": {hour: at10, price: "100", asOf: at10, ok: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			price, asOf, ok := history.PriceAt(test.hour)

			assert.Equal(t, test.ok, ok)

			if !test.ok {
				return
			}

			assert.True(t, dec(t, test.price).Equal(price))
			assert.Equal(t, test.asOf, asOf)
			assert.Equal(t, test.stale, asOf.Before(test.hour))
		})
	}
}
