package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"time"
)

var ErrResponse = errors.New("internal server ")

type PriceExchange struct {
	Symbol   string  `json:"symbol"`
	Exchange string  `json:"exchange"`
	Value    float64 `json:"value"`
}

type ServerResponse struct {
	Time time.Time `json:"time"`
}

type Candlestick struct {
	StartTime  time.Time `json:"start_time"`
	OpenPrice  float64   `json:"open_price"`
	HighPrice  float64   `json:"high_price"`
	LowPrice   float64   `json:"low_price"`
	ClosePrice float64   `json:"close_price"`
	Volume     float64   `json:"volume"`
}

type Indicator struct {
	Datetime time.Time `json:"datetime"`
	Value    float64   `json:"value"`
}

func DefaultClient() *Client {
	c, err := NewClient("http://localhost:8085")
	if err != nil {
		zap.L().Panic("error init crypto_polymath client", zap.Error(err))
	}
	return c
}

type Client struct {
	httpClient *http.Client
	hostUrl    *url.URL
}

func NewClient(host string) (*Client, error) {
	hostUrl, err := url.Parse(host)
	if err != nil {
		return nil, errors.Wrap(err, "parse host")
	}
	return &Client{
		hostUrl:    hostUrl,
		httpClient: http.DefaultClient,
	}, nil
}

func (c *Client) SetHttpClient(httpClient *http.Client) {
	c.httpClient = httpClient
}

func (c *Client) Server(ctx context.Context) (ServerResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/server", c.hostUrl.String()),
		nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ServerResponse{}, errors.Wrap(err, "send request")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusInternalServerError {
		return ServerResponse{}, ErrResponse
	}
	var result ServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ServerResponse{}, err
	}

	return result, nil
}

func (c *Client) PricesByExchange(ctx context.Context, exchange string) ([]PriceExchange, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/prices/exchange/%s", c.hostUrl.String(), exchange),
		nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "send request")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusInternalServerError {
		return nil, ErrResponse
	}
	var result []PriceExchange
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) PricesBySymbol(ctx context.Context, symbol string) ([]PriceExchange, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/prices/symbol/%s", c.hostUrl.String(), symbol),
		nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "send request")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusInternalServerError {
		return nil, ErrResponse
	}
	var result []PriceExchange
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) Prices(ctx context.Context, exchange, symbol string) (PriceExchange, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/price/%s/%s", c.hostUrl.String(), exchange, symbol),
		nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return PriceExchange{}, errors.Wrap(err, "send request")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusInternalServerError {
		return PriceExchange{}, ErrResponse
	}
	var result PriceExchange
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return PriceExchange{}, err
	}

	return result, nil
}

func (c *Client) Candlesticks(ctx context.Context, exchange, symbol, unit string, interval int) ([]Candlestick, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/candlestick/%s/%s/%s/%d", c.hostUrl.String(), exchange, symbol, unit, interval),
		nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "send request")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusInternalServerError {
		return nil, ErrResponse
	}
	var result []Candlestick
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) Indicators(ctx context.Context, exchange, symbol, unit string, interval int, name string, depth int) ([]Indicator, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/indicator/%s/%s/%s/%d/%s/%d", c.hostUrl.String(), exchange, symbol, unit, interval, name, depth),
		nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "send request")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusInternalServerError {
		return nil, ErrResponse
	}
	var result []Indicator
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
