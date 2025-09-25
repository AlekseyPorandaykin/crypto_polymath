package repository

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/pkg/errors"
)

var _ candlestick.Repository = (*CandlestickRepository)(nil)

type CandlestickRepository struct {
	storage candlestick.Repository
	cache   candlestick.Repository
}

func NewCandlestickRepository(storage, cache candlestick.Repository) candlestick.Repository {
	return &CandlestickRepository{storage: storage, cache: cache}
}

func (c *CandlestickRepository) Save(ctx context.Context, data ...candlestick.StorageDTO) error {
	_ = c.cache.Save(ctx, data...)
	if err := c.storage.Save(ctx, data...); err != nil {
		return errors.Wrap(err, "from storage")
	}
	return nil
}

func (c *CandlestickRepository) Last(ctx context.Context, exchange, symbol, unit string, interval, limit, offset int) ([]candlestick.StorageDTO, error) {
	dataCache, err := c.cache.Last(ctx, exchange, symbol, unit, interval, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "from cache")
	}
	if len(dataCache) > 0 {
		if candlestick.IsPrevCandle(dataCache[0]) {
			return dataCache, nil
		}
	}
	dataStorage, err := c.storage.Last(ctx, exchange, symbol, unit, interval, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "from storage")
	}
	if len(dataStorage) == 0 {
		return nil, nil
	}
	_ = c.updateCache(ctx, exchange, symbol, unit, interval)
	return dataStorage, nil
}

func (c *CandlestickRepository) DeleteOldRows(ctx context.Context, exchange, symbol, unit string, interval int, to time.Time) error {
	if err := c.storage.DeleteOldRows(ctx, exchange, symbol, unit, interval, to); err != nil {
		return errors.Wrap(err, "from storage")
	}
	if err := c.cache.DeleteOldRows(ctx, exchange, symbol, unit, interval, to); err != nil {
		return errors.Wrap(err, "from cache")
	}
	return nil
}

func (c *CandlestickRepository) DeletePrevRows(ctx context.Context, exchange, symbol, unit string, interval int, to time.Time) error {
	if err := c.storage.DeletePrevRows(ctx, exchange, symbol, unit, interval, to); err != nil {
		return errors.Wrap(err, "from storage")
	}
	if err := c.cache.DeletePrevRows(ctx, exchange, symbol, unit, interval, to); err != nil {
		return errors.Wrap(err, "from cache")
	}
	return nil
}

func (c *CandlestickRepository) LastToDate(ctx context.Context, exchange, symbol, unit string, interval, limit int, to time.Time) ([]candlestick.StorageDTO, error) {
	dataCache, err := c.cache.LastToDate(ctx, exchange, symbol, unit, interval, limit, to)
	if err != nil {
		return nil, errors.Wrap(err, "from cache")
	}
	if dataCache != nil {
		return dataCache, nil
	}
	dataStorage, err := c.storage.LastToDate(ctx, exchange, symbol, unit, interval, limit, to)
	if err != nil {
		return nil, errors.Wrap(err, "from storage")
	}
	if len(dataStorage) == 0 {
		return nil, nil
	}
	_ = c.updateCache(ctx, exchange, symbol, unit, interval)
	return dataStorage, nil
}

func (c *CandlestickRepository) FromDate(ctx context.Context, exchange, symbol, unit string, interval, limit int, to time.Time) ([]candlestick.StorageDTO, error) {
	dataCache, err := c.cache.FromDate(ctx, exchange, symbol, unit, interval, limit, to)
	if err != nil {
		return nil, errors.Wrap(err, "from cache")
	}
	if dataCache != nil {
		return dataCache, nil
	}
	dataStorage, err := c.storage.FromDate(ctx, exchange, symbol, unit, interval, limit, to)
	if err != nil {
		return nil, errors.Wrap(err, "from storage")
	}
	if len(dataStorage) == 0 {
		return nil, nil
	}
	_ = c.updateCache(ctx, exchange, symbol, unit, interval)
	return dataStorage, nil
}

func (c *CandlestickRepository) ListUniq(ctx context.Context) ([]candlestick.UniqDTO, error) {
	dataCache, err := c.cache.ListUniq(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "from cache")
	}
	if dataCache != nil {
		return dataCache, nil
	}
	dataStorage, err := c.storage.ListUniq(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "from storage")
	}
	return dataStorage, nil
}

func (c *CandlestickRepository) updateCache(ctx context.Context, exchange, symbol, unit string, interval int) error {
	data, err := c.storage.Last(ctx, exchange, symbol, unit, interval, 500, 0)
	if err != nil {
		return err
	}
	return c.cache.Save(ctx, data...)
}
