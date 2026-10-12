package domain_test

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/domain"
)

func hour(t *testing.T, s string) time.Time {
	t.Helper()

	at, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)

	return at.UTC()
}

func dec(t *testing.T, s string) decimal.Decimal {
	t.Helper()

	value, err := decimal.NewFromString(s)
	require.NoError(t, err)

	return value
}

func mustDec(s string) decimal.Decimal {
	value, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}

	return value
}

func buy(id domain.OperationID, at time.Time, quantity string) domain.Operation {
	return domain.Operation{ID: id, Type: domain.OperationBuy, ExecutedAt: at, Quantity: mustDec(quantity)}
}

func sell(id domain.OperationID, at time.Time, quantity string) domain.Operation {
	return domain.Operation{ID: id, Type: domain.OperationSell, ExecutedAt: at, Quantity: mustDec(quantity)}
}

func TestBalanceTimeline_BalanceAt(t *testing.T) {
	t0 := hour(t, "2026-01-01T10:00:00Z")
	t1 := hour(t, "2026-01-02T10:00:00Z")
	t2 := hour(t, "2026-01-03T10:00:00Z")

	fee := dec(t, "0.5")

	timeline := domain.NewBalanceTimeline([]domain.Operation{
		buy(2, t2, "3"),
		buy(1, t0, "1"),
		{ID: 3, Type: domain.OperationSell, ExecutedAt: t1, Quantity: dec(t, "2"), FeeAmount: &fee},
	})

	tests := map[string]struct {
		at   time.Time
		want string
	}{
		"before all operations": {at: t0.Add(-time.Hour), want: "0"},
		"at first buy":          {at: t0, want: "1"},
		"after sell with fee":   {at: t1, want: "-1"},
		"after later buy":       {at: t2, want: "2"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.True(t, dec(t, test.want).Equal(timeline.BalanceAt(test.at)))
		})
	}
}

func TestBalanceTimeline_FeeNeverAffectsBalance(t *testing.T) {
	at := hour(t, "2026-01-01T10:00:00Z")
	fee := dec(t, "100")

	timeline := domain.NewBalanceTimeline([]domain.Operation{
		{ID: 1, Type: domain.OperationBuy, ExecutedAt: at, Quantity: dec(t, "5"), FeeAmount: &fee},
	})

	assert.True(t, dec(t, "5").Equal(timeline.BalanceAt(at)))
}

func TestBalanceTimeline_BalanceMayBeNegative(t *testing.T) {
	at := hour(t, "2026-01-01T10:00:00Z")

	timeline := domain.NewBalanceTimeline([]domain.Operation{sell(1, at, "2")})

	assert.True(t, dec(t, "-2").Equal(timeline.BalanceAt(at)))
}

func TestBalanceTimeline_HoldingIntervals(t *testing.T) {
	t0 := hour(t, "2026-01-01T10:30:00Z")
	t1 := hour(t, "2026-01-02T09:15:00Z")
	until := hour(t, "2026-01-05T00:00:00Z")

	tests := map[string]struct {
		operations []domain.Operation
		want       []domain.HoldingInterval
	}{
		"open position ends at current hour": {
			operations: []domain.Operation{buy(1, t0, "1")},
			want:       []domain.HoldingInterval{{From: t0.Truncate(time.Hour), To: until}},
		},
		"closed position has inclusive hour bounds": {
			operations: []domain.Operation{buy(1, t0, "1"), sell(2, t1, "1")},
			want: []domain.HoldingInterval{
				{From: t0.Truncate(time.Hour), To: t1.Truncate(time.Hour)},
			},
		},
		"zero balance at start produces no interval": {
			operations: []domain.Operation{sell(1, t0, "1")},
			want:       []domain.HoldingInterval{{From: t0.Truncate(time.Hour), To: until}},
		},
		"buy and sell at the same hour produce a one-hour interval": {
			operations: []domain.Operation{buy(1, t0, "1"), sell(2, t0, "1")},
			want:       []domain.HoldingInterval{{From: t0.Truncate(time.Hour), To: t0.Truncate(time.Hour)}},
		},
		"separate intervals": {
			operations: []domain.Operation{
				buy(1, t0, "1"),
				sell(2, t1, "1"),
				buy(3, hour(t, "2026-01-03T10:00:00Z"), "2"),
			},
			want: []domain.HoldingInterval{
				{From: t0.Truncate(time.Hour), To: t1.Truncate(time.Hour)},
				{From: hour(t, "2026-01-03T10:00:00Z"), To: until},
			},
		},
		"re-entering the same hour merges intervals": {
			operations: []domain.Operation{
				buy(1, t0, "1"),
				sell(2, t1, "1"),
				buy(3, t1, "1"),
			},
			want: []domain.HoldingInterval{{From: t0.Truncate(time.Hour), To: until}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			timeline := domain.NewBalanceTimeline(test.operations)

			assert.Equal(t, test.want, timeline.HoldingIntervals(until))
		})
	}
}

func TestBalanceTimeline_HoldingIntervals_BackdatedOperationExtends(t *testing.T) {
	first := hour(t, "2026-01-10T10:00:00Z")
	backdated := hour(t, "2026-01-01T08:00:00Z")
	until := hour(t, "2026-01-15T00:00:00Z")

	timeline := domain.NewBalanceTimeline([]domain.Operation{
		buy(1, first, "1"),
		buy(2, backdated, "1"),
	})

	assert.Equal(t, []domain.HoldingInterval{
		{From: backdated.Truncate(time.Hour), To: until},
	}, timeline.HoldingIntervals(until))
}
