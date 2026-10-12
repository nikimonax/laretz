package coingecko

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/pkg/plugin"
)

func ms(t time.Time) string {
	return strconv.FormatInt(t.UnixMilli(), 10)
}

func TestPlanRequests(t *testing.T) {
	now := time.Date(2026, 10, 12, 15, 30, 45, 0, time.UTC)
	cutoffHourly := hourFloor(now.Add(-hourlyMaxAge))
	cutoffDaily := hourFloor(now.Add(-dailyMaxAge))

	cases := []struct {
		name       string
		quoteRange plugin.QuoteRange
		expected   []chartRequest
	}{
		{
			name:       "skips a range ending at the daily cutoff",
			quoteRange: plugin.QuoteRange{ExternalID: "bitcoin", From: cutoffDaily.Add(-24 * time.Hour), To: cutoffDaily},
			expected:   []chartRequest{},
		},
		{
			name:       "serves only hourly inside the hourly window",
			quoteRange: plugin.QuoteRange{ExternalID: "bitcoin", From: now.Add(-24 * time.Hour), To: now},
			expected: []chartRequest{
				{externalID: "bitcoin", from: now.Add(-24 * time.Hour), to: now, granularity: plugin.GranularityHour},
			},
		},
		{
			name:       "serves only daily between the cutoffs",
			quoteRange: plugin.QuoteRange{ExternalID: "bitcoin", From: cutoffDaily.Add(24 * time.Hour), To: cutoffHourly.Add(-24 * time.Hour)},
			expected: []chartRequest{
				{externalID: "bitcoin", from: cutoffDaily.Add(24 * time.Hour), to: cutoffHourly.Add(-24 * time.Hour), granularity: plugin.GranularityDay},
			},
		},
		{
			name:       "splits a wide range into daily and hourly",
			quoteRange: plugin.QuoteRange{ExternalID: "bitcoin", From: cutoffDaily.Add(-24 * time.Hour), To: now},
			expected: []chartRequest{
				{externalID: "bitcoin", from: cutoffDaily, to: cutoffHourly.Add(-time.Hour), granularity: plugin.GranularityDay},
				{externalID: "bitcoin", from: cutoffHourly, to: now, granularity: plugin.GranularityHour},
			},
		},
		{
			name:       "omits an empty daily slice",
			quoteRange: plugin.QuoteRange{ExternalID: "bitcoin", From: cutoffHourly, To: now},
			expected: []chartRequest{
				{externalID: "bitcoin", from: cutoffHourly, to: now, granularity: plugin.GranularityHour},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			requests := planRequests([]plugin.QuoteRange{tt.quoteRange}, now)

			assert.Equal(t, tt.expected, requests)
		})
	}
}

func TestQuoteProviderFetch(t *testing.T) {
	ctx := context.Background()

	t.Run("fetches hourly points and drops the ones outside the range", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Hour)
		from := now.Add(-24 * time.Hour)
		to := now

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "hourly", r.URL.Query().Get("interval"))
			_, _ = w.Write([]byte(`{"prices":[
				[` + ms(from.Add(-time.Hour)) + `,"1"],
				[` + ms(from) + `,"2"],
				[` + ms(to) + `,"3"],
				[` + ms(to.Add(time.Hour)) + `,"4"]
			]}`))
		}))
		t.Cleanup(server.Close)

		provider := NewQuoteProvider(NewClient(server.Client(), server.URL, ""))

		quotes, err := provider.Fetch(ctx, []plugin.QuoteRange{{ExternalID: "bitcoin", From: from, To: to}})

		require.NoError(t, err)
		require.Len(t, quotes, 2)
		assert.Equal(t, "bitcoin", quotes[0].ExternalID)
		assert.Equal(t, from, quotes[0].At)
		assert.Equal(t, "2", quotes[0].Price)
		assert.Equal(t, plugin.GranularityHour, quotes[0].Granularity)
		assert.Equal(t, to, quotes[1].At)
		assert.Equal(t, "3", quotes[1].Price)
	})

	t.Run("places daily points at the day start hour", func(t *testing.T) {
		from := time.Now().UTC().Truncate(24 * time.Hour).Add(-100 * 24 * time.Hour)
		to := from.Add(95 * time.Hour)
		point := from.Add(48*time.Hour + 13*time.Hour + 37*time.Minute)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "daily", r.URL.Query().Get("interval"))
			_, _ = w.Write([]byte(`{"prices":[
				[` + ms(from.Add(-time.Hour)) + `,"1"],
				[` + ms(point) + `,"7"]
			]}`))
		}))
		t.Cleanup(server.Close)

		provider := NewQuoteProvider(NewClient(server.Client(), server.URL, ""))

		quotes, err := provider.Fetch(ctx, []plugin.QuoteRange{{ExternalID: "bitcoin", From: from, To: to}})

		require.NoError(t, err)
		require.Len(t, quotes, 1)
		assert.Equal(t, from.Add(48*time.Hour), quotes[0].At)
		assert.Equal(t, "7", quotes[0].Price)
		assert.Equal(t, plugin.GranularityDay, quotes[0].Granularity)
	})

	t.Run("stops on an upstream error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "boom", http.StatusInternalServerError)
		}))
		t.Cleanup(server.Close)

		provider := NewQuoteProvider(NewClient(server.Client(), server.URL, ""))
		now := time.Now().UTC().Truncate(time.Hour)

		quotes, err := provider.Fetch(ctx, []plugin.QuoteRange{{
			ExternalID: "bitcoin",
			From:       now.Add(-24 * time.Hour),
			To:         now,
		}})

		require.ErrorContains(t, err, "unexpected status")
		assert.Empty(t, quotes)
	})
}
