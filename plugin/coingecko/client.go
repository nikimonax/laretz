package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	DefaultBaseURL = "https://api.coingecko.com/api/v3"
	demoKeyHeader  = "x-cg-demo-api-key"

	vsCurrencyUSD = "usd"

	maxResponseBytes = 64 << 20
)

type Client struct {
	http    *http.Client
	baseURL string
	apiKey  string
}

func NewClient(client *http.Client, baseURL string, apiKey string) *Client {
	return &Client{
		http:    client,
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

type Coin struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

func (client *Client) CoinsList(ctx context.Context) ([]Coin, error) {
	endpoint, err := url.JoinPath(client.baseURL, "coins/list")

	if err != nil {
		return nil, fmt.Errorf("coins list: %w", err)
	}

	request, err := newRequest(ctx, client, endpoint)

	if err != nil {
		return nil, fmt.Errorf("coins list: %w", err)
	}

	return doRequest[[]Coin](client, request, "coins list")
}

// ChartPoint is one price observation of the chart data: Time is a moment in
// UTC, Price is the decimal text of the price in USD.
type ChartPoint struct {
	Time  time.Time
	Price string
}

type MarketChart struct {
	Prices []ChartPoint
}

type marketChartResponse struct {
	Prices [][]json.Number `json:"prices"`
}

// MarketChartRange requests the historical chart data of a coin within the
// given time range with the given interval ("hourly" or "daily").
func (client *Client) MarketChartRange(ctx context.Context, coinID string, from time.Time, to time.Time, interval string) (MarketChart, error) {
	endpoint, err := url.JoinPath(client.baseURL, "coins", url.PathEscape(coinID), "market_chart/range")

	if err != nil {
		return MarketChart{}, fmt.Errorf("market chart %s: %w", coinID, err)
	}

	query := url.Values{
		"vs_currency": {vsCurrencyUSD},
		"from":        {strconv.FormatInt(from.Unix(), 10)},
		"to":          {strconv.FormatInt(to.Unix(), 10)},
	}

	if interval != "" {
		query.Set("interval", interval)
	}

	request, err := newRequest(ctx, client, endpoint+"?"+query.Encode())

	if err != nil {
		return MarketChart{}, fmt.Errorf("market chart %s: %w", coinID, err)
	}

	payload, err := doRequest[marketChartResponse](client, request, fmt.Sprintf("market chart %s", coinID))

	if err != nil {
		return MarketChart{}, err
	}

	chart := MarketChart{Prices: make([]ChartPoint, 0, len(payload.Prices))}

	for _, pair := range payload.Prices {
		if len(pair) < 2 {
			continue
		}

		timestamp, err := strconv.ParseFloat(pair[0].String(), 64)

		if err != nil {
			return MarketChart{}, fmt.Errorf("market chart %s: timestamp: %w", coinID, err)
		}

		chart.Prices = append(chart.Prices, ChartPoint{Time: time.UnixMilli(int64(timestamp)).UTC(), Price: pair[1].String()})
	}

	return chart, nil
}

func newRequest(ctx context.Context, client *Client, endpoint string) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)

	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", "application/json")

	if client.apiKey != "" {
		request.Header.Set(demoKeyHeader, client.apiKey)
	}

	return request, nil
}

func doRequest[T any](client *Client, request *http.Request, name string) (T, error) {
	response, err := client.http.Do(request)

	if err != nil {
		return *new(T), fmt.Errorf("%s: %w", name, err)
	}

	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return *new(T), fmt.Errorf("%s: unexpected status %s", name, response.Status)
	}

	var payload T

	if err := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes)).Decode(&payload); err != nil {
		return *new(T), fmt.Errorf("%s: decode: %w", name, err)
	}

	return payload, nil
}
