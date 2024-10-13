package adapters

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/pkg/errors"
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
	data, err := c.candlestickService.SequenceCandlesticksToDate(ctx, exchange, symbol, string(unit), interval, limit, datetime)
	if err != nil {
		return nil, errors.Wrap(err, "get sequence candlesticks to date")
	}
	return data, nil
}

func (c *CandlestickAdapter) SequenceCandlesticks(
	ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int,
) ([]domain.Candlestick, error) {
	data, err := c.candlestickService.SequenceCandlesticks(ctx, exchange, symbol, unit, interval, limit)
	if err != nil {
		return nil, errors.Wrap(err, "get sequence candlesticks")
	}
	return data, nil
}

func (c *CandlestickAdapter) NextCandlesticks(
	ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int, datetime time.Time,
) ([]domain.Candlestick, error) {
	now := time.Now().In(time.UTC)
	extremeDatetime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 1, 0, time.UTC)
	if extremeDatetime.Before(datetime) || extremeDatetime.Equal(datetime) {
		return nil, nil
	}
	data, err := c.candlestickService.CandlesticksFromDate(ctx, exchange, symbol, string(unit), interval, limit, datetime)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *CandlestickAdapter) FirstCandlestick(
	ctx context.Context, exchange, symbol string, unit domain.Unit, interval int, offset int,
) (*domain.Candlestick, error) {
	data, err := c.candlestickService.SequenceCandlesticks(ctx, exchange, symbol, unit, interval, 100)
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
