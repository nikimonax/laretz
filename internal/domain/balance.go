package domain

import (
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

// HoldingInterval is a maximal span of time during which the Balance of an
// Asset in a Portfolio is non-zero. From and To are inclusive hour boundaries
// (UTC); an interval still open at the moment of computation ends at the
// current hour.
type HoldingInterval struct {
	From time.Time
	To   time.Time
}

// BalanceTimeline derives the Balance of an Asset in a Portfolio at an
// arbitrary moment from its Operations: Buy adds Quantity, Sell subtracts it.
// A Fee never affects the Balance; the Balance may be negative.
type BalanceTimeline struct {
	operations []Operation
}

func NewBalanceTimeline(operations []Operation) *BalanceTimeline {
	sorted := make([]Operation, len(operations))

	copy(sorted, operations)

	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].ExecutedAt.Equal(sorted[j].ExecutedAt) {
			return sorted[i].ID < sorted[j].ID
		}

		return sorted[i].ExecutedAt.Before(sorted[j].ExecutedAt)
	})

	return &BalanceTimeline{operations: sorted}
}

// BalanceAt returns the Balance at the given moment: the signed sum of all
// Operations executed up to and including that moment.
func (timeline *BalanceTimeline) BalanceAt(at time.Time) decimal.Decimal {
	balance := decimal.Zero

	for _, operation := range timeline.operations {
		if operation.ExecutedAt.After(at) {
			break
		}

		balance = balance.Add(operation.signedQuantity())
	}

	return balance
}

// HoldingIntervals returns the maximal spans of non-zero Balance. The hours of
// the boundary Operations are included; an interval still open at until ends
// at the current hour.
func (timeline *BalanceTimeline) HoldingIntervals(until time.Time) []HoldingInterval {
	var (
		intervals []HoldingInterval
		open      *HoldingInterval
		balance   = decimal.Zero
	)

	for _, operation := range timeline.operations {
		next := balance.Add(operation.signedQuantity())

		switch {
		case balance.IsZero() && !next.IsZero():
			open = &HoldingInterval{From: hourBoundary(operation.ExecutedAt)}
		case !balance.IsZero() && next.IsZero():
			open.To = hourBoundary(operation.ExecutedAt)
			intervals = append(intervals, *open)
			open = nil
		}

		balance = next
	}

	if open != nil {
		if end := hourBoundary(until); !end.Before(open.From) {
			open.To = end
			intervals = append(intervals, *open)
		}
	}

	return mergeHoldingIntervals(intervals)
}

func (operation Operation) signedQuantity() decimal.Decimal {
	if operation.Type == OperationSell {
		return operation.Quantity.Neg()
	}

	return operation.Quantity
}

func hourBoundary(t time.Time) time.Time {
	return t.UTC().Truncate(time.Hour)
}

// mergeHoldingIntervals collapses intervals that touch or overlap at hour
// granularity; the input is already ordered by From.
func mergeHoldingIntervals(intervals []HoldingInterval) []HoldingInterval {
	merged := make([]HoldingInterval, 0, len(intervals))

	for _, interval := range intervals {
		last := len(merged) - 1

		if last >= 0 && !interval.From.After(merged[last].To) {
			if interval.To.After(merged[last].To) {
				merged[last].To = interval.To
			}

			continue
		}

		merged = append(merged, interval)
	}

	return merged
}
