package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	DefaultBaseURL = "https://api.coingecko.com/api/v3"
	demoKeyHeader  = "x-cg-demo-api-key"

	maxResponseBytes = 64 << 20
)

type Coin struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

type Client struct {
	http    *http.Client
	baseURL string
	apiKey  string
}

func (client *Client) CoinsList(ctx context.Context) ([]Coin, error) {
	endpoint, err := url.JoinPath(client.baseURL, "coins/list")

	if err != nil {
		return nil, fmt.Errorf("coins list: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)

	if err != nil {
		return nil, fmt.Errorf("coins list: %w", err)
	}

	request.Header.Set("Accept", "application/json")

	if client.apiKey != "" {
		request.Header.Set(demoKeyHeader, client.apiKey)
	}

	response, err := client.http.Do(request)

	if err != nil {
		return nil, fmt.Errorf("coins list: %w", err)
	}

	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coins list: unexpected status %s", response.Status)
	}

	var coins []Coin

	if err := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes)).Decode(&coins); err != nil {
		return nil, fmt.Errorf("coins list: decode: %w", err)
	}

	return coins, nil
}

func NewClient(client *http.Client, baseURL string, apiKey string) *Client {
	return &Client{
		http:    client,
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}
