package domain

import (
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

// PriceHistory is the set of PricePoints recorded for one ExternalAsset,
// ordered by At.
type PriceHistory []PricePoint

// PriceAt resolves the price in effect at the given hour: the PricePoint of
// the hour itself, or the last known PricePoint before it, carried forward
// for display only (ADR-0002). ok is false when no PricePoint exists at or
// before the hour.
func (history PriceHistory) PriceAt(hour time.Time) (price decimal.Decimal, asOf time.Time, ok bool) {
	target := hour.UTC().Truncate(time.Hour)

	for _, point := range history {
		at := point.At.UTC().Truncate(time.Hour)

		if at.After(target) {
			break
		}

		price, asOf, ok = point.PriceUSD, at, true
	}

	return price, asOf, ok
}

// ValuedAsset is an Asset prepared for valuation: its Operations as a
// BalanceTimeline and the PriceHistory of its primary ExternalAsset.
type ValuedAsset struct {
	Asset    Asset
	Timeline *BalanceTimeline
	Prices   PriceHistory
}

// Value is the display-time worth of holdings at a moment: the sum over the
// held Assets of Balance × Price, with the display flags of the prices used.
type Value struct {
	Amount        decimal.Decimal
	Stale         bool // at least one price was carried forward from an earlier hour
	MissingPrices bool // at least one held Asset has no price
}

// ValuedPosition is a held Asset valued at a moment: its Balance and the
// price resolved through the primary link, with the display-time staleness
// flags. An Asset without a resolvable price is not Priced and contributes
// nothing to the value.
type ValuedPosition struct {
	Asset    Asset
	Balance  decimal.Decimal
	PriceUSD decimal.Decimal
	AsOf     time.Time
	Priced   bool
	Stale    bool
	Value    decimal.Decimal
}

// HourValue is the value of the holdings at one hour of the chart axis.
// Hours inside no HoldingInterval are not Held and carry no Amount.
type HourValue struct {
	Hour          time.Time
	Held          bool
	Amount        decimal.Decimal
	Stale         bool
	MissingPrices bool
}

// Valuation values the Assets held in a Portfolio for display: at any moment
// every held Asset contributes Balance × Price, where the price is the
// PricePoint of the hour or the last known one carried forward (ADR-0002).
// A negative Balance reduces the value.
type Valuation struct {
	assets []ValuedAsset
}

func NewValuation(assets []ValuedAsset) *Valuation {
	return &Valuation{assets: assets}
}

// At values the holdings at the given moment and returns the total value
// together with the valued positions of the Assets held at that moment,
// ordered by symbol.
func (valuation *Valuation) At(at time.Time) (Value, []ValuedPosition) {
	hour := at.UTC().Truncate(time.Hour)

	value := Value{Amount: decimal.Zero}
	positions := make([]ValuedPosition, 0, len(valuation.assets))

	for _, asset := range valuation.assets {
		balance := asset.Timeline.BalanceAt(at)

		if balance.IsZero() {
			continue
		}

		position := ValuedPosition{Asset: asset.Asset, Balance: balance}

		if price, asOf, ok := asset.Prices.PriceAt(hour); ok {
			position.Priced = true
			position.PriceUSD = price
			position.AsOf = asOf
			position.Stale = asOf.Before(hour)
			position.Value = balance.Mul(price)

			value.Amount = value.Amount.Add(position.Value)
			value.Stale = value.Stale || position.Stale
		} else {
			value.MissingPrices = true
		}

		positions = append(positions, position)
	}

	sort.SliceStable(positions, func(i, j int) bool {
		if positions[i].Asset.Symbol != positions[j].Asset.Symbol {
			return positions[i].Asset.Symbol < positions[j].Asset.Symbol
		}

		return positions[i].Asset.ID < positions[j].Asset.ID
	})

	return value, positions
}

// Intervals returns the union of the Assets' HoldingIntervals, merged at
// hour granularity.
func (valuation *Valuation) Intervals(until time.Time) []HoldingInterval {
	intervals := make([]HoldingInterval, 0, len(valuation.assets))

	for _, asset := range valuation.assets {
		intervals = append(intervals, asset.Timeline.HoldingIntervals(until)...)
	}

	sort.SliceStable(intervals, func(i, j int) bool {
		return intervals[i].From.Before(intervals[j].From)
	})

	return mergeHoldingIntervals(intervals)
}

// Series values every hour spanned by the holdings' active intervals: hours
// inside an interval carry the value of the holdings at that hour; hours
// between intervals carry none.
func (valuation *Valuation) Series(until time.Time) []HourValue {
	intervals := valuation.Intervals(until)

	if len(intervals) == 0 {
		return nil
	}

	first, last := intervals[0], intervals[len(intervals)-1]
	hours := make([]HourValue, 0, int(last.To.Sub(first.From).Hours())+1)

	for hour := first.From; !hour.After(last.To); hour = hour.Add(time.Hour) {
		if !holdsAny(intervals, hour) {
			hours = append(hours, HourValue{Hour: hour})
			continue
		}

		value, _ := valuation.At(hour)
		hours = append(hours, HourValue{
			Hour:          hour,
			Held:          true,
			Amount:        value.Amount,
			Stale:         value.Stale,
			MissingPrices: value.MissingPrices,
		})
	}

	return hours
}

func holdsAny(intervals []HoldingInterval, hour time.Time) bool {
	for _, interval := range intervals {
		if !hour.Before(interval.From) && !hour.After(interval.To) {
			return true
		}
	}

	return false
}
