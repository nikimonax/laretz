package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nikimonax/laretz/internal/domain"
)

func pricePoint(externalAssetID domain.ExternalAssetID, at time.Time, granularity domain.Granularity) domain.PricePoint {
	return domain.PricePoint{ExternalAssetID: externalAssetID, At: at, Granularity: granularity}
}

func TestQuoteCoverage_Covers(t *testing.T) {
	at := hour(t, "2026-01-01T10:00:00Z")

	coverage := domain.NewQuoteCoverage([]domain.PricePoint{
		pricePoint(1, at, domain.GranularityHour),
		pricePoint(1, hour(t, "2026-01-02T00:00:00Z"), domain.GranularityDay),
	})

	tests := map[string]struct {
		hour time.Time
		want bool
	}{
		"hourly point covers its own hour":        {hour: at, want: true},
		"non-UTC time truncates to the same hour": {hour: at.In(time.FixedZone("x", 7200)), want: true},
		"hourly point does not cover other hours": {hour: at.Add(time.Hour), want: false},
		"daily point covers the whole day":        {hour: hour(t, "2026-01-02T23:00:00Z"), want: true},
		"daily point does not cover next day":     {hour: hour(t, "2026-01-03T00:00:00Z"), want: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, coverage.Covers(test.hour))
		})
	}
}

func TestCoveragePlanner_Plan(t *testing.T) {
	asset := domain.ExternalAsset{SourceID: "coingecko", ExternalID: "bitcoin"}

	interval := domain.HoldingInterval{
		From: hour(t, "2026-01-01T10:00:00Z"),
		To:   hour(t, "2026-01-01T13:00:00Z"),
	}

	tests := map[string]struct {
		intervals []domain.HoldingInterval
		coverage  domain.QuoteCoverage
		want      []domain.QuoteRange
	}{
		"fully covered yields no ranges": {
			intervals: []domain.HoldingInterval{interval},
			coverage: domain.NewQuoteCoverage([]domain.PricePoint{
				pricePoint(1, interval.From, domain.GranularityHour),
				pricePoint(1, interval.From.Add(time.Hour), domain.GranularityHour),
				pricePoint(1, interval.To.Add(-time.Hour), domain.GranularityHour),
				pricePoint(1, interval.To, domain.GranularityHour),
			}),
			want: nil,
		},
		"uncovered interval yields one range with inclusive bounds": {
			intervals: []domain.HoldingInterval{interval},
			want: []domain.QuoteRange{{
				SourceID:   asset.SourceID,
				ExternalID: asset.ExternalID,
				From:       interval.From,
				To:         interval.To,
			}},
		},
		"gap inside coverage yields one range": {
			intervals: []domain.HoldingInterval{interval},
			coverage: domain.NewQuoteCoverage([]domain.PricePoint{
				pricePoint(1, interval.From, domain.GranularityHour),
				pricePoint(1, interval.To, domain.GranularityHour),
			}),
			want: []domain.QuoteRange{{
				SourceID:   asset.SourceID,
				ExternalID: asset.ExternalID,
				From:       interval.From.Add(time.Hour),
				To:         interval.To.Add(-time.Hour),
			}},
		},
		"daily point covers the whole day, no refetch": {
			intervals: []domain.HoldingInterval{interval},
			coverage: domain.NewQuoteCoverage([]domain.PricePoint{
				pricePoint(1, hour(t, "2026-01-01T00:00:00Z"), domain.GranularityDay),
			}),
			want: nil,
		},
		"backdated interval before coverage extends the plan backwards": {
			intervals: []domain.HoldingInterval{
				{From: hour(t, "2025-12-31T20:00:00Z"), To: hour(t, "2025-12-31T22:00:00Z")},
				interval,
			},
			coverage: domain.NewQuoteCoverage([]domain.PricePoint{
				pricePoint(1, interval.From, domain.GranularityHour),
			}),
			want: []domain.QuoteRange{
				{
					SourceID:   asset.SourceID,
					ExternalID: asset.ExternalID,
					From:       hour(t, "2025-12-31T20:00:00Z"),
					To:         hour(t, "2025-12-31T22:00:00Z"),
				},
				{
					SourceID:   asset.SourceID,
					ExternalID: asset.ExternalID,
					From:       interval.From.Add(time.Hour),
					To:         interval.To,
				},
			},
		},
		"adjacent intervals touching at hour granularity merge": {
			intervals: []domain.HoldingInterval{
				{From: hour(t, "2026-01-01T10:00:00Z"), To: hour(t, "2026-01-01T10:00:00Z")},
				{From: hour(t, "2026-01-01T11:00:00Z"), To: hour(t, "2026-01-01T11:00:00Z")},
			},
			want: []domain.QuoteRange{{
				SourceID:   asset.SourceID,
				ExternalID: asset.ExternalID,
				From:       hour(t, "2026-01-01T10:00:00Z"),
				To:         hour(t, "2026-01-01T11:00:00Z"),
			}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, domain.CoveragePlanner{}.Plan(asset, test.intervals, test.coverage))
		})
	}
}
