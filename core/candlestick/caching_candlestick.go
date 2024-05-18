package candlestick

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"time"
)

type cachingCandlestick struct {
	cache   Candlestick
	service Candlestick
}

func NewCachingCandlestick(cache, service Candlestick) Candlestick {
	return &cachingCandlestick{}
}

func (c *cachingCandlestick) AddLoader(exchange string, loader ExchangeLoader) {
	c.service.AddLoader(exchange, loader)
}

func (c *cachingCandlestick) LoadCandlesticksMinutes(
	ctx context.Context, exchange, symbol string, minutes int,
) ([]domain.Candlestick, error) {
	data, err := c.service.LoadCandlesticksMinutes(ctx, exchange, symbol, minutes)
	if err != nil {
		return nil, err
	}
	_ = c.cache.Save(ctx, data...)
	return data, nil
}

func (c *cachingCandlestick) LoadCandlesticksHours(
	ctx context.Context, exchange, symbol string, hours int,
) ([]domain.Candlestick, error) {
	data, err := c.service.LoadCandlesticksHours(ctx, exchange, symbol, hours)
	if err != nil {
		return nil, err
	}
	_ = c.cache.Save(ctx, data...)
	return data, nil
}

func (c *cachingCandlestick) DeleteOldRows(ctx context.Context, oldValueLimit int) error {
	return c.service.DeleteOldRows(ctx, oldValueLimit)
}

func (c *cachingCandlestick) CandlesticksToDate(ctx context.Context, exchange, symbol, unit string, minutes, limit int, to time.Time) ([]domain.Candlestick, error) {
	return c.service.CandlesticksToDate(ctx, exchange, symbol, unit, minutes, limit, to)
}

func (c *cachingCandlestick) CandlesticksFromDate(ctx context.Context, exchange, symbol, unit string, minutes, limit int, to time.Time) ([]domain.Candlestick, error) {
	return c.service.CandlesticksFromDate(ctx, exchange, symbol, unit, minutes, limit, to)
}

func (c *cachingCandlestick) LoadCandlesticksDay(
	ctx context.Context, exchange, symbol string,
) ([]domain.Candlestick, error) {
	data, err := c.service.LoadCandlesticksDay(ctx, exchange, symbol)
	if err != nil {
		return nil, err
	}
	_ = c.cache.Save(ctx, data...)
	return data, nil
}

func (c *cachingCandlestick) LoadCandlesticksWeek(
	ctx context.Context, exchange, symbol string,
) ([]domain.Candlestick, error) {
	data, err := c.service.LoadCandlesticksWeek(ctx, exchange, symbol)
	if err != nil {
		return nil, err
	}
	_ = c.cache.Save(ctx, data...)
	return data, nil
}

func (c *cachingCandlestick) LoadCandlesticksMonth(
	ctx context.Context, exchange, symbol string,
) ([]domain.Candlestick, error) {
	data, err := c.service.LoadCandlesticksMonth(ctx, exchange, symbol)
	if err != nil {
		return nil, err
	}
	_ = c.cache.Save(ctx, data...)
	return data, nil
}

func (c *cachingCandlestick) Save(ctx context.Context, candlestick ...domain.Candlestick) error {
	if err := c.service.Save(ctx, candlestick...); err != nil {
		return err
	}
	if err := c.cache.Save(ctx, candlestick...); err != nil {
		return err
	}
	return nil
}

func (c *cachingCandlestick) Candlestick(
	ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int,
) ([]domain.Candlestick, error) {
	return c.service.Candlestick(ctx, exchange, symbol, unit, interval, limit)
}

func (c *cachingCandlestick) CandlesticksMinutes(
	ctx context.Context, exchange, symbol string, minutes, limit int,
) ([]domain.Candlestick, error) {
	dataCache, err := c.cache.CandlesticksMinutes(ctx, exchange, symbol, minutes, limit)
	if err != nil {
		return nil, err
	}
	if len(dataCache) > 0 {
		return dataCache, nil
	}
	data, err := c.service.CandlesticksMinutes(ctx, exchange, symbol, minutes, limit)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *cachingCandlestick) CandlesticksHours(
	ctx context.Context, exchange, symbol string, hours, limit int,
) ([]domain.Candlestick, error) {
	dataCache, err := c.cache.CandlesticksHours(ctx, exchange, symbol, hours, limit)
	if err != nil {
		return nil, err
	}
	if len(dataCache) > 0 {
		return dataCache, nil
	}
	data, err := c.service.CandlesticksHours(ctx, exchange, symbol, hours, limit)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *cachingCandlestick) CandlesticksDay(
	ctx context.Context, exchange, symbol string, limit int,
) ([]domain.Candlestick, error) {
	dataCache, err := c.cache.CandlesticksDay(ctx, exchange, symbol, limit)
	if err != nil {
		return nil, err
	}
	if len(dataCache) > 0 {
		return dataCache, nil
	}
	data, err := c.service.CandlesticksDay(ctx, exchange, symbol, limit)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *cachingCandlestick) CandlesticksWeek(
	ctx context.Context, exchange, symbol string, limit int,
) ([]domain.Candlestick, error) {
	dataCache, err := c.cache.CandlesticksWeek(ctx, exchange, symbol, limit)
	if err != nil {
		return nil, err
	}
	if len(dataCache) > 0 {
		return dataCache, nil
	}
	data, err := c.service.CandlesticksWeek(ctx, exchange, symbol, limit)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *cachingCandlestick) CandlesticksMonth(
	ctx context.Context, exchange, symbol string, limit int,
) ([]domain.Candlestick, error) {
	dataCache, err := c.cache.CandlesticksMonth(ctx, exchange, symbol, limit)
	if err != nil {
		return nil, err
	}
	if len(dataCache) > 0 {
		return dataCache, nil
	}
	data, err := c.service.CandlesticksMonth(ctx, exchange, symbol, limit)
	if err != nil {
		return nil, err
	}
	return data, nil
}
