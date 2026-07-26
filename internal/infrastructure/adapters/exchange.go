package adapters

import (
	"context"
	"fmt"

	"github.com/AlekseyPorandaykin/crypto-exchanges/client"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/go-kit/pkg/util"
)

var _ candlestick.ExchangeLoader = (*Exchange)(nil)
var _ price.ExchangeLoader = (*Exchange)(nil)
var _ exchange.ExternalLoader = (*Exchange)(nil)

type Exchange struct {
	c client.ExchangeClient
}

func NewExchange(c client.ExchangeClient) *Exchange {
	return &Exchange{c: c}
}

func (e *Exchange) Name() string {
	return e.c.ExchangeName()
}

func (e *Exchange) SymbolInfo(ctx context.Context) ([]exchange.SymbolInfoDTO, error) {
	data, err := e.c.SymbolInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting symbol info: %w", err)
	}
	return util.ModifySlice(data, func(item client.SymbolInfo) exchange.SymbolInfoDTO {
		return exchange.SymbolInfoDTO{
			Symbol:          item.Symbol,
			Exchange:        item.Exchange,
			BaseAsset:       item.BaseAsset,
			QuoteAsset:      item.QuoteAsset,
			Category:        exchange.SymbolCategory(item.Category),
			FundingRate:     item.FundingRate,
			NextFundingTime: item.NextFundingTime,
		}
	}), nil
}

func (e *Exchange) Prices(ctx context.Context) ([]price.ExchangeDTO, error) {
	data, err := e.c.Prices(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting prices: %w", err)
	}
	return util.ModifySlice(data, func(item client.Price) price.ExchangeDTO {
		return price.ExchangeDTO{
			Symbol:    item.Symbol,
			Exchange:  item.Exchange,
			Value:     item.Value,
			CreatedAt: item.CreatedAt,
		}
	}), nil
}

func (e *Exchange) Price(ctx context.Context, symbol string) (price.ExchangeDTO, error) {
	data, err := e.Prices(ctx)
	if err != nil {
		return price.ExchangeDTO{}, err
	}
	for _, item := range data {
		if item.Symbol == symbol {
			return item, nil
		}
	}
	return price.ExchangeDTO{}, nil
}

func (e *Exchange) LastMinuteCandlesticks(ctx context.Context, symbol string, minutes int) ([]candlestick.ExchangeDTO, error) {
	data, err := e.c.LastMinuteCandlesticks(ctx, symbol, minutes)
	if err != nil {
		return nil, fmt.Errorf("getting last minute candlesticks: %w", err)
	}
	return util.ModifySlice(data, func(item client.Candlestick) candlestick.ExchangeDTO {
		return candlestick.ExchangeDTO{
			StartTime:  item.StartTime,
			OpenPrice:  item.OpenPrice,
			HighPrice:  item.HighPrice,
			LowPrice:   item.LowPrice,
			ClosePrice: item.ClosePrice,
			Volume:     item.Volume,
		}
	}), nil
}

func (e *Exchange) LastHourCandlesticks(ctx context.Context, symbol string, hours int) ([]candlestick.ExchangeDTO, error) {
	data, err := e.c.LastHourCandlesticks(ctx, symbol, hours)
	if err != nil {
		return nil, fmt.Errorf("getting last hours candlesticks: %w", err)
	}
	return util.ModifySlice(data, func(item client.Candlestick) candlestick.ExchangeDTO {
		return candlestick.ExchangeDTO{
			StartTime:  item.StartTime,
			OpenPrice:  item.OpenPrice,
			HighPrice:  item.HighPrice,
			LowPrice:   item.LowPrice,
			ClosePrice: item.ClosePrice,
			Volume:     item.Volume,
		}
	}), nil
}

func (e *Exchange) LastDayCandlesticks(ctx context.Context, symbol string) ([]candlestick.ExchangeDTO, error) {
	data, err := e.c.LastDayCandlesticks(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("getting last day candlesticks: %w", err)
	}
	return util.ModifySlice(data, func(item client.Candlestick) candlestick.ExchangeDTO {
		return candlestick.ExchangeDTO{
			StartTime:  item.StartTime,
			OpenPrice:  item.OpenPrice,
			HighPrice:  item.HighPrice,
			LowPrice:   item.LowPrice,
			ClosePrice: item.ClosePrice,
			Volume:     item.Volume,
		}
	}), nil
}

func (e *Exchange) LastWeekCandlesticks(ctx context.Context, symbol string) ([]candlestick.ExchangeDTO, error) {
	data, err := e.c.LastWeekCandlesticks(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("getting last week candlesticks: %w", err)
	}
	return util.ModifySlice(data, func(item client.Candlestick) candlestick.ExchangeDTO {
		return candlestick.ExchangeDTO{
			StartTime:  item.StartTime,
			OpenPrice:  item.OpenPrice,
			HighPrice:  item.HighPrice,
			LowPrice:   item.LowPrice,
			ClosePrice: item.ClosePrice,
			Volume:     item.Volume,
		}
	}), nil
}

func (e *Exchange) LastMonthCandlesticks(ctx context.Context, symbol string) ([]candlestick.ExchangeDTO, error) {
	data, err := e.c.LastMonthCandlesticks(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("getting last month candlesticks: %w", err)
	}
	return util.ModifySlice(data, func(item client.Candlestick) candlestick.ExchangeDTO {
		return candlestick.ExchangeDTO{
			StartTime:  item.StartTime,
			OpenPrice:  item.OpenPrice,
			HighPrice:  item.HighPrice,
			LowPrice:   item.LowPrice,
			ClosePrice: item.ClosePrice,
			Volume:     item.Volume,
		}
	}), nil
}
