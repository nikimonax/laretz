package plugin

import (
	"context"
	"time"

	"github.com/go-logr/logr"
)

type Asset struct {
	ID     string
	Symbol string
	Name   string
}

type AssetProvider interface {
	Provide(ctx context.Context) ([]Asset, error)
}

type Granularity string

const (
	GranularityHour Granularity = "hour"
	GranularityDay  Granularity = "day"
)

// QuoteRange is a continuous span of hours with inclusive boundaries, for
// which a quote source is asked for prices.
type QuoteRange struct {
	ExternalID string
	From       time.Time
	To         time.Time
}

// Quote is a price as returned by a quote source. At is an hour boundary in
// UTC; Price is the decimal text of the price in USD.
type Quote struct {
	ExternalID  string
	At          time.Time
	Price       string
	Granularity Granularity
}

type QuoteProvider interface {
	Fetch(ctx context.Context, ranges []QuoteRange) ([]Quote, error)
}

type App interface {
	Logger() logr.Logger

	GetConfig(cfg any) error
	SetConfig(cfg any) error

	RegisterAssetProvider(provider AssetProvider) error
	RegisterQuoteProvider(provider QuoteProvider) error
}

type PluginID = string

type Plugin interface {
	ID() PluginID
	Init(app App) error
}
