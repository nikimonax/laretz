package coingecko

import (
	"context"
	"time"

	"github.com/nikimonax/laretz/pkg/plugin"
)

const (
	// The CoinGecko demo key serves hourly chart data only for the last
	// ~90 days and daily data only up to ~365 days back.
	hourlyMaxAge = 90 * 24 * time.Hour
	dailyMaxAge  = 365 * 24 * time.Hour

	// Pause between consecutive requests to stay within the demo rate limit.
	requestPause = 2 * time.Second

	intervalHourly = "hourly"
	intervalDaily  = "daily"
)

type QuoteProvider struct {
	client *Client
}

// Fetch implements [plugin.QuoteProvider].
func (provider *QuoteProvider) Fetch(ctx context.Context, ranges []plugin.QuoteRange) ([]plugin.Quote, error) {
	requests := planRequests(ranges, time.Now().UTC())

	quotes := make([]plugin.Quote, 0, len(requests))

	for i, request := range requests {
		if i > 0 {
			if err := pause(ctx); err != nil {
				return nil, err
			}
		}

		fetched, err := provider.fetch(ctx, request)

		if err != nil {
			return nil, err
		}

		quotes = append(quotes, fetched...)
	}

	return quotes, nil
}

type chartRequest struct {
	externalID  string
	from        time.Time
	to          time.Time
	granularity plugin.Granularity
}

// planRequests splits every range by the granularity the demo tier serves:
// hourly for the last ~90 days, daily before that, and nothing older than
// ~365 days.
func planRequests(ranges []plugin.QuoteRange, now time.Time) []chartRequest {
	cutoffHourly := hourFloor(now.Add(-hourlyMaxAge))
	cutoffDaily := hourFloor(now.Add(-dailyMaxAge))

	requests := make([]chartRequest, 0, 2*len(ranges))

	for _, quoteRange := range ranges {
		if !quoteRange.To.After(cutoffDaily) {
			continue
		}

		from := quoteRange.From

		if from.Before(cutoffDaily) {
			from = cutoffDaily
		}

		dailyTo := quoteRange.To

		if !dailyTo.Before(cutoffHourly) {
			dailyTo = cutoffHourly.Add(-time.Hour)
		}

		if !dailyTo.Before(from) {
			requests = append(requests, chartRequest{
				externalID:  quoteRange.ExternalID,
				from:        from,
				to:          dailyTo,
				granularity: plugin.GranularityDay,
			})
		}

		hourlyFrom := quoteRange.From

		if hourlyFrom.Before(cutoffHourly) {
			hourlyFrom = cutoffHourly
		}

		if !quoteRange.To.Before(hourlyFrom) {
			requests = append(requests, chartRequest{
				externalID:  quoteRange.ExternalID,
				from:        hourlyFrom,
				to:          quoteRange.To,
				granularity: plugin.GranularityHour,
			})
		}
	}

	return requests
}

func (provider *QuoteProvider) fetch(ctx context.Context, request chartRequest) ([]plugin.Quote, error) {
	chart, err := provider.client.MarketChartRange(ctx, request.externalID, request.from, request.to, intervalParam(request.granularity))

	if err != nil {
		return nil, err
	}

	quotes := make([]plugin.Quote, 0, len(chart.Prices))

	for _, point := range chart.Prices {
		at := hourFloor(point.Time)

		if at.Before(request.from) || at.After(request.to) {
			continue
		}

		// Daily points are placed at the day's start hour (ADR-0002).
		if request.granularity == plugin.GranularityDay {
			at = at.Truncate(24 * time.Hour)
		}

		quotes = append(quotes, plugin.Quote{
			ExternalID:  request.externalID,
			At:          at,
			Price:       point.Price,
			Granularity: request.granularity,
		})
	}

	return quotes, nil
}

func pause(ctx context.Context) error {
	timer := time.NewTimer(requestPause)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func hourFloor(t time.Time) time.Time {
	return t.UTC().Truncate(time.Hour)
}

func intervalParam(granularity plugin.Granularity) string {
	if granularity == plugin.GranularityDay {
		return intervalDaily
	}

	return intervalHourly
}

func NewQuoteProvider(client *Client) plugin.QuoteProvider {
	return &QuoteProvider{client: client}
}
