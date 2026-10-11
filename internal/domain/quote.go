package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Granularity string

const (
	GranularityHour Granularity = "hour"
	GranularityDay  Granularity = "day"
)

// PricePoint is a price of an ExternalAsset at an hour boundary, expressed in
// the Quote Currency. Daily points sit at the day's start hour; gaps between
// points are legitimate.
type PricePoint struct {
	ExternalAssetID ExternalAssetID
	At              time.Time
	PriceUSD        decimal.Decimal
	Granularity     Granularity
}

// QuoteRange is a continuous span of hours with inclusive boundaries, required
// for an ExternalAsset but missing from its QuoteCoverage.
type QuoteRange struct {
	SourceID   SourceID
	ExternalID string
	From       time.Time
	To         time.Time
}

// FetchedQuote is a quote as returned by a quote source, identified by the
// source and the source's external ID. Only hours really served by the source
// appear here.
type FetchedQuote struct {
	SourceID    SourceID
	ExternalID  string
	At          time.Time
	PriceUSD    decimal.Decimal
	Granularity Granularity
}
