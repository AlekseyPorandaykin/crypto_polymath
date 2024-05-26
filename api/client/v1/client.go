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

var ErrResponse = errors.New("internal server")
var ErrNotSupportedResponse = errors.New("don't supported response")
var ErrNotFound = errors.New("not found")

type ErrClient struct {
	err     error
	details string
}

func (e ErrClient) Error() string {
	return e.err.Error()
}

func (e ErrClient) Details() string {
	return e.details
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

func (c *Client) Server(ctx context.Context) (ServerInfoResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/server", c.hostUrl.String()),
		nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ServerInfoResponse{}, errors.Wrap(err, "send request")
	}
	var result ServerInfoResponse

	if err := c.parseResponse(resp, &result); err != nil {
		return ServerInfoResponse{}, err
	}

	return result, nil
}

func (c *Client) PricesByExchange(ctx context.Context, exchange string) (PricesResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/prices/exchange/%s", c.hostUrl.String(), exchange),
		nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "send request")
	}
	var result PricesResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}
func (c *Client) PricesBySymbol(ctx context.Context, symbol string) (PricesResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/prices/symbol/%s", c.hostUrl.String(), symbol),
		nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "send request")
	}
	var result PricesResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}
func (c *Client) Prices(ctx context.Context, exchange, symbol string) (PriceResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/price/%s/%s", c.hostUrl.String(), exchange, symbol),
		nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return PriceResponse{}, errors.Wrap(err, "send request")
	}
	var result PriceResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return PriceResponse{}, err
	}

	return result, nil
}

func (c *Client) ExchangerSymbol(ctx context.Context, exchange, symbol string) (*SymbolInfoResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/exchange/%s/%s", c.hostUrl.String(), exchange, symbol),
		nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "send request")
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	var result SymbolInfoResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DayCandlesticks(ctx context.Context, exchange, symbol string) (CandlesticksResponse, error) {
	return c.candlesticks(ctx, exchange, symbol, GetCandlestickExchangeSymbolUnitIntervalParamsUnitD, 1)
}
func (c *Client) HourCandlesticks(ctx context.Context, exchange, symbol string, interval int) (CandlesticksResponse, error) {
	return c.candlesticks(ctx, exchange, symbol, GetCandlestickExchangeSymbolUnitIntervalParamsUnitH, interval)
}
func (c *Client) MinuteCandlesticks(ctx context.Context, exchange, symbol string, interval int) (CandlesticksResponse, error) {
	return c.candlesticks(ctx, exchange, symbol, GetCandlestickExchangeSymbolUnitIntervalParamsUnitM, interval)
}
func (c *Client) MonthCandlesticks(ctx context.Context, exchange, symbol string) (CandlesticksResponse, error) {
	return c.candlesticks(ctx, exchange, symbol, GetCandlestickExchangeSymbolUnitIntervalParamsUnitM1, 1)
}
func (c *Client) WeekCandlesticks(ctx context.Context, exchange, symbol string) (CandlesticksResponse, error) {
	return c.candlesticks(ctx, exchange, symbol, GetCandlestickExchangeSymbolUnitIntervalParamsUnitW, 1)
}
func (c *Client) candlesticks(
	ctx context.Context,
	exchange string,
	symbol string,
	unit GetCandlestickExchangeSymbolUnitIntervalParamsUnit,
	interval int,
) (CandlesticksResponse, error) {
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
	var result CandlesticksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) EmaDayIndicators(ctx context.Context, exchange, symbol string, interval, depth int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitD,
		interval,
		EMA,
		depth,
	)
}
func (c *Client) EmaHourIndicators(ctx context.Context, exchange, symbol string, interval, depth int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitH,
		interval,
		EMA,
		depth,
	)
}
func (c *Client) EmaMinuteIndicators(ctx context.Context, exchange, symbol string, interval, depth int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitM,
		interval,
		EMA,
		depth,
	)
}
func (c *Client) EmaMonthIndicators(ctx context.Context, exchange, symbol string, interval, depth int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitM1,
		interval,
		EMA,
		depth,
	)
}
func (c *Client) EmaWeekIndicators(ctx context.Context, exchange, symbol string, interval, depth int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitW,
		interval,
		EMA,
		depth,
	)
}

func (c *Client) MaDayIndicators(ctx context.Context, exchange, symbol string, interval, depth int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitD,
		interval,
		MA,
		depth,
	)
}
func (c *Client) MaHourIndicators(ctx context.Context, exchange, symbol string, interval, depth int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitH,
		interval,
		MA,
		depth,
	)
}
func (c *Client) MaMinuteIndicators(ctx context.Context, exchange, symbol string, interval, depth int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitM,
		interval,
		MA,
		depth,
	)
}
func (c *Client) MaMonthIndicators(ctx context.Context, exchange, symbol string, interval, depth int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitM1,
		interval,
		MA,
		depth,
	)
}
func (c *Client) MaWeekIndicators(ctx context.Context, exchange, symbol string, interval, depth int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitW,
		interval,
		MA,
		depth,
	)
}

func (c *Client) TrendDayIndicators(ctx context.Context, exchange, symbol string, interval, depth int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitD,
		interval,
		Trend,
		depth,
	)
}
func (c *Client) TrendHourIndicators(ctx context.Context, exchange, symbol string, interval, depth int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitH,
		interval,
		Trend,
		depth,
	)
}
func (c *Client) TrendMinuteIndicators(ctx context.Context, exchange, symbol string, interval, depth int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitM,
		interval,
		Trend,
		depth,
	)
}
func (c *Client) TrendMonthIndicators(ctx context.Context, exchange, symbol string, interval, depth int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitM1,
		interval,
		Trend,
		depth,
	)
}
func (c *Client) TrendWeekIndicators(ctx context.Context, exchange, symbol string, interval, depth int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitW,
		interval,
		Trend,
		depth,
	)
}

func (c *Client) TypeCandleDayIndicators(ctx context.Context, exchange, symbol string, interval int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitD,
		interval,
		TypeCandle,
		1,
	)
}
func (c *Client) TypeCandleHourIndicators(ctx context.Context, exchange, symbol string, interval int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitH,
		interval,
		TypeCandle,
		1,
	)
}
func (c *Client) TypeCandleMinuteIndicators(ctx context.Context, exchange, symbol string, interval int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitM,
		interval,
		TypeCandle,
		1,
	)
}
func (c *Client) TypeCandleMonthIndicators(ctx context.Context, exchange, symbol string, interval int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitM1,
		interval,
		TypeCandle,
		1,
	)
}
func (c *Client) TypeCandleWeekIndicators(ctx context.Context, exchange, symbol string, interval int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitW,
		interval,
		TypeCandle,
		1,
	)
}

func (c *Client) VolatilityCandlePercentDayIndicators(ctx context.Context, exchange string, symbol string, interval int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitD,
		interval,
		VolatilityCandlePercent,
		1,
	)
}
func (c *Client) VolatilityCandlePercentHourIndicators(ctx context.Context, exchange string, symbol string, interval int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitH,
		interval,
		VolatilityCandlePercent,
		1,
	)
}
func (c *Client) VolatilityCandlePercentMinuteIndicators(ctx context.Context, exchange string, symbol string, interval int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitM,
		interval,
		VolatilityCandlePercent,
		1,
	)
}
func (c *Client) VolatilityCandlePercentMonthIndicators(ctx context.Context, exchange string, symbol string, interval int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitM1,
		interval,
		VolatilityCandlePercent,
		1,
	)
}
func (c *Client) VolatilityCandlePercentWeekIndicators(ctx context.Context, exchange string, symbol string, interval int) (IndicatorResponse, error) {
	return c.indicators(
		ctx,
		exchange,
		symbol,
		GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnitW,
		interval,
		VolatilityCandlePercent,
		1,
	)
}

func (c *Client) indicators(
	ctx context.Context,
	exchange string,
	symbol string,
	unit GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsUnit,
	interval int,
	name GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsName,
	depth int,
) (IndicatorResponse, error) {
	if err := checkDepth(depth); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(
			"%s/api/indicator/%s/%s/%s/%d/%s/%d",
			c.hostUrl.String(),
			exchange,
			symbol,
			unit,
			interval,
			name,
			depth,
		),
		nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "send request")
	}
	var result IndicatorResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) parseResponse(resp *http.Response, dest interface{}) error {
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusOK {
		return json.NewDecoder(resp.Body).Decode(dest)
	}
	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	errClient := ErrClient{err: ErrNotSupportedResponse}
	if resp.StatusCode == http.StatusInternalServerError {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			errClient.details = err.Error()
			return errClient
		}
		errClient.details = errResp.Message
	}
	return errClient
}

func checkDepth(depth int) error {
	switch GetIndicatorExchangeSymbolUnitIntervalNameDepthParamsDepth(depth) {
	case N1, N10, N20, N50:
		return nil
	default:
		return errors.New("don't support depth")
	}
}
