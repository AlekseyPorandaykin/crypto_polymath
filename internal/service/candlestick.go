package service

import (
	"context"
	"errors"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/duke-git/lancet/v2/slice"
	"time"
)

type CandlestickAdapter struct {
	candlestickService candlestick.Candlestick
}

func NewCandlestickAdapter(candlestickService candlestick.Candlestick) *CandlestickAdapter {
	return &CandlestickAdapter{candlestickService: candlestickService}
}

func (c *CandlestickAdapter) LastCandlesticks(
	ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int, datetime time.Time,
) ([]domain.Candlestick, error) {
	data, err := c.candlestickService.CandlesticksToDate(ctx, exchange, symbol, string(unit), interval, limit, datetime)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *CandlestickAdapter) NextCandlesticks(
	ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int, datetime time.Time,
) ([]domain.Candlestick, error) {
	data, err := c.candlestickService.CandlesticksFromDate(ctx, exchange, symbol, string(unit), interval, limit, datetime)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *CandlestickAdapter) FirstCandlestick(
	ctx context.Context, exchange, symbol string, unit domain.Unit, interval int, offset int,
) (*domain.Candlestick, error) {
	data, err := c.unitCandlesticks(ctx, exchange, symbol, unit, interval, 100)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	slice.SortBy[domain.Candlestick](data, func(a, b domain.Candlestick) bool {
		return a.StartTime.Before(b.StartTime)
	})
	if offset > 0 && offset < len(data) {
		return &data[offset-1], nil
	}

	return &data[0], nil
}

func (c *CandlestickAdapter) unitCandlesticks(ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int) ([]domain.Candlestick, error) {
	switch unit {
	case domain.MinuteUnit:
		return c.candlestickService.CandlesticksMinutes(ctx, exchange, symbol, interval, limit)
	case domain.HourUnit:
		return c.candlestickService.CandlesticksHours(ctx, exchange, symbol, interval, limit)
	case domain.DayUnit:
		if interval != 1 {
			return nil, errors.New("don't support interval")
		}
		return c.candlestickService.CandlesticksDay(ctx, exchange, symbol, limit)
	case domain.WeekUnit:
		if interval != 1 {
			return nil, errors.New("don't support interval")
		}
		return c.candlestickService.CandlesticksWeek(ctx, exchange, symbol, limit)
	case domain.MonthUnit:
		if interval != 1 {
			return nil, errors.New("don't support interval")
		}
		return c.candlestickService.CandlesticksMonth(ctx, exchange, symbol, limit)
	default:
		return nil, errors.New("don't support unit")
	}
}
