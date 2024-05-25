package exchange

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/bybit/v5"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/bybit/v5/request"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	core_exchange "github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

const BybitExchange = "bybit"

type Bybit struct {
	client *v5.Client
}

func NewByBit(client *v5.Client) *Bybit {
	return &Bybit{client: client}
}

func (c *Bybit) LastMinuteCandlesticks(ctx context.Context, symbol string, minutes int) ([]candlestick.ExchangeDTO, error) {
	return c.lastCandlesticks(ctx, symbol, strconv.Itoa(minutes))
}

func (c *Bybit) LastHourCandlesticks(ctx context.Context, symbol string, hours int) ([]candlestick.ExchangeDTO, error) {
	minutes := hours * 60
	return c.lastCandlesticks(ctx, symbol, strconv.Itoa(minutes))
}

func (c *Bybit) LastDayCandlesticks(ctx context.Context, symbol string) ([]candlestick.ExchangeDTO, error) {
	return c.lastCandlesticks(ctx, symbol, "D")
}

func (c *Bybit) LastWeekCandlesticks(ctx context.Context, symbol string) ([]candlestick.ExchangeDTO, error) {
	return c.lastCandlesticks(ctx, symbol, "W")
}

func (c *Bybit) LastMonthCandlesticks(ctx context.Context, symbol string) ([]candlestick.ExchangeDTO, error) {
	return c.lastCandlesticks(ctx, symbol, "M")
}

func (c *Bybit) lastCandlesticks(ctx context.Context, symbol, interval string) ([]candlestick.ExchangeDTO, error) {
	resp, err := c.client.MarketGetKline(ctx, request.MarketGetKlineParam{
		Symbol:   symbol,
		Interval: interval,
		Limit:    100,
	})
	if err != nil {
		return nil, errors.Wrap(err, "get MarketGetKline")
	}
	if !resp.IsOk() {
		return nil, errors.New("incorrect response lastCandlesticks")
	}
	data := resp.Result.Candlesticks()
	result := make([]candlestick.ExchangeDTO, 0, len(data))
	for _, item := range data {
		startTimeMs, err := strconv.Atoi(item.StartTime)
		if err != nil {
			continue
		}
		result = append(result, candlestick.ExchangeDTO{
			StartTime:  time.UnixMilli(int64(startTimeMs)).In(time.UTC),
			OpenPrice:  item.OpenPrice,
			HighPrice:  item.HighPrice,
			LowPrice:   item.LowPrice,
			ClosePrice: item.ClosePrice,
			Volume:     item.Volume,
		})
	}

	return result, nil
}

func (c *Bybit) Prices(ctx context.Context) ([]price.ExchangeDTO, error) {
	result := make([]price.ExchangeDTO, 0, 500)
	var tickerResp v5.TickerResponse
	err := backoff.Retry(func() error {
		var err error
		tickerResp, err = c.client.MarketSpotTicker(ctx)
		if err != nil {
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, errors.Wrap(err, "error get price from bybit")
	}
	if !tickerResp.IsOk() {
		return nil, fmt.Errorf("incorrect response from bybit: %s", tickerResp.Message)
	}
	currentTime := time.UnixMilli(tickerResp.Time)
	if currentTime.Year() != time.Now().Year() {
		currentTime = time.Now()
	}
	for _, data := range tickerResp.Result.List {
		result = append(result, price.ExchangeDTO{
			Symbol:    data.Symbol,
			Exchange:  "bybit",
			Value:     data.LastPrice,
			CreatedAt: currentTime,
		})
	}
	return result, nil
}

func (c *Bybit) Price(ctx context.Context, symbol string) (price.ExchangeDTO, error) {
	return price.ExchangeDTO{}, nil
}

func (c *Bybit) SymbolInfo(ctx context.Context) ([]core_exchange.SymbolInfoDTO, error) {
	instrumentInfo, err := c.client.MarketInstrumentsInfo(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]core_exchange.SymbolInfoDTO, 0, len(instrumentInfo.Result.List))
	for _, item := range instrumentInfo.Result.List {
		result = append(result, core_exchange.SymbolInfoDTO{
			Symbol:     item.Symbol,
			Exchange:   BybitExchange,
			BaseAsset:  item.BaseCoin,
			QuoteAsset: item.QuoteCoin,
		})
	}

	return result, nil
}
