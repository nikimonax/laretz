package coingecko

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoinsList(t *testing.T) {
	ctx := context.Background()

	t.Run("fetches the catalog and sends the demo key header", func(t *testing.T) {
		var (
			path    string
			accept  string
			demoKey string
		)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path = r.URL.Path
			accept = r.Header.Get("Accept")
			demoKey = r.Header.Get(demoKeyHeader)
			_, _ = w.Write([]byte(`[
				{"id":"bitcoin","symbol":"btc","name":"Bitcoin"},
				{"id":"ethereum","symbol":"eth","name":"Ethereum"}
			]`))
		}))
		t.Cleanup(server.Close)

		client := NewClient(server.Client(), server.URL, "test-key")

		coins, err := client.CoinsList(ctx)

		require.NoError(t, err)
		assert.Equal(t, "/coins/list", path)
		assert.Equal(t, "application/json", accept)
		assert.Equal(t, "test-key", demoKey)
		assert.Equal(t, []Coin{
			{ID: "bitcoin", Symbol: "btc", Name: "Bitcoin"},
			{ID: "ethereum", Symbol: "eth", Name: "Ethereum"},
		}, coins)
	})

	t.Run("omits the demo key header without a key", func(t *testing.T) {
		var demoKey string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			demoKey = r.Header.Get(demoKeyHeader)
			_, _ = w.Write([]byte(`[]`))
		}))
		t.Cleanup(server.Close)

		client := NewClient(server.Client(), server.URL, "")

		coins, err := client.CoinsList(ctx)

		require.NoError(t, err)
		assert.Empty(t, coins)
		assert.Empty(t, demoKey)
	})

	t.Run("rejects a non-200 response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "boom", http.StatusTeapot)
		}))
		t.Cleanup(server.Close)

		client := NewClient(server.Client(), server.URL, "")

		_, err := client.CoinsList(ctx)

		require.ErrorContains(t, err, "unexpected status")
	})

	t.Run("rejects a broken payload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("not json"))
		}))
		t.Cleanup(server.Close)

		client := NewClient(server.Client(), server.URL, "")

		_, err := client.CoinsList(ctx)

		require.ErrorContains(t, err, "decode")
	})
}

func TestMarketChartRange(t *testing.T) {
	ctx := context.Background()

	t.Run("requests the range and parses the price points", func(t *testing.T) {
		var (
			path  string
			query url.Values
		)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path = r.URL.Path
			query = r.URL.Query()
			_, _ = w.Write([]byte(`{"prices":[[1700000000000,"27000.5"],[1700003600000,"27100"]]}`))
		}))
		t.Cleanup(server.Close)

		client := NewClient(server.Client(), server.URL, "test-key")
		from := time.Unix(1700000000, 0).UTC()
		to := time.Unix(1700003600, 0).UTC()

		chart, err := client.MarketChartRange(ctx, "bitcoin", from, to, "hourly")

		require.NoError(t, err)
		assert.Equal(t, "/coins/bitcoin/market_chart/range", path)
		assert.Equal(t, "usd", query.Get("vs_currency"))
		assert.Equal(t, "1700000000", query.Get("from"))
		assert.Equal(t, "1700003600", query.Get("to"))
		assert.Equal(t, "hourly", query.Get("interval"))
		assert.Equal(t, []ChartPoint{
			{Time: time.UnixMilli(1700000000000).UTC(), Price: "27000.5"},
			{Time: time.UnixMilli(1700003600000).UTC(), Price: "27100"},
		}, chart.Prices)
	})

	t.Run("omits the interval when unset", func(t *testing.T) {
		var query url.Values
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query = r.URL.Query()
			_, _ = w.Write([]byte(`{"prices":[]}`))
		}))
		t.Cleanup(server.Close)

		client := NewClient(server.Client(), server.URL, "")

		chart, err := client.MarketChartRange(ctx, "bitcoin", time.Unix(1700000000, 0).UTC(), time.Unix(1700003600, 0).UTC(), "")

		require.NoError(t, err)
		assert.Empty(t, chart.Prices)
		assert.Empty(t, query["interval"])
	})

	t.Run("skips short points", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"prices":[[1700000000000],[1700003600000,"27100"]]}`))
		}))
		t.Cleanup(server.Close)

		client := NewClient(server.Client(), server.URL, "")

		chart, err := client.MarketChartRange(ctx, "bitcoin", time.Unix(1700000000, 0).UTC(), time.Unix(1700003600, 0).UTC(), "")

		require.NoError(t, err)
		require.Len(t, chart.Prices, 1)
		assert.Equal(t, "27100", chart.Prices[0].Price)
	})

	t.Run("rejects an unparsable timestamp", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"prices":[[1e400,"1"]]}`))
		}))
		t.Cleanup(server.Close)

		client := NewClient(server.Client(), server.URL, "")

		_, err := client.MarketChartRange(ctx, "bitcoin", time.Unix(1700000000, 0).UTC(), time.Unix(1700003600, 0).UTC(), "")

		require.ErrorContains(t, err, "timestamp")
	})
}
