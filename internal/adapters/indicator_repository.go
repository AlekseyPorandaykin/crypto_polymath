package adapters

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/pkg/errors"
	"time"
)

var _ indicator.Repository = (*IndicatorRepository)(nil)

type IndicatorRepository struct {
	storage indicator.Repository
	cache   indicator.Repository
}

func NewIndicatorRepository(storage, cache indicator.Repository) indicator.Repository {
	return &IndicatorRepository{storage: storage, cache: cache}
}

func (i *IndicatorRepository) Save(ctx context.Context, data ...indicator.StorageDTO) error {
	if err := i.storage.Save(ctx, data...); err != nil {
		return errors.Wrap(err, "save to storage")
	}
	if err := i.storage.Save(ctx, data...); err != nil {
		return errors.Wrap(err, "save to cache")
	}
	return nil
}

func (i *IndicatorRepository) Find(
	ctx context.Context, exchange, symbol, unit string, interval int, datetime time.Time, name string, depth int,
) (*indicator.StorageDTO, error) {
	cacheData, err := i.cache.Find(ctx, exchange, symbol, unit, interval, datetime, name, depth)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from cache")
	}
	if cacheData != nil {
		return cacheData, nil
	}
	storageData, err := i.storage.Find(ctx, exchange, symbol, unit, interval, datetime, name, depth)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from cache")
	}
	i.updateCache(ctx, exchange, symbol, unit, interval, name, depth)
	return storageData, nil
}

func (i *IndicatorRepository) List(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth, limit, offset int,
) ([]indicator.StorageDTO, error) {
	cacheData, err := i.cache.List(ctx, exchange, symbol, unit, interval, name, depth, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from cache")
	}
	if cacheData != nil {
		return cacheData, nil
	}
	storageData, err := i.storage.List(ctx, exchange, symbol, unit, interval, name, depth, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from cache")
	}
	i.updateCache(ctx, exchange, symbol, unit, interval, name, depth)
	return storageData, nil
}

func (i *IndicatorRepository) Last(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth int,
) (*indicator.StorageDTO, error) {
	cacheData, err := i.cache.Last(ctx, exchange, symbol, unit, interval, name, depth)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from cache")
	}
	if cacheData != nil {
		return cacheData, nil
	}
	storageData, err := i.storage.Last(ctx, exchange, symbol, unit, interval, name, depth)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from cache")
	}
	i.updateCache(ctx, exchange, symbol, unit, interval, name, depth)
	return storageData, nil
}

func (i *IndicatorRepository) DeleteOldRows(
	ctx context.Context, symbol, exchangeName, unit string, interval int, name string, depth int, to time.Time,
) error {
	_ = i.cache.DeleteOldRows(ctx, symbol, exchangeName, unit, interval, name, depth, to)
	return i.storage.DeleteOldRows(ctx, symbol, exchangeName, unit, interval, name, depth, to)
}

func (i *IndicatorRepository) ListUniq(ctx context.Context) ([]indicator.UniqDTO, error) {
	cacheData, err := i.cache.ListUniq(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from cache")
	}
	if cacheData != nil {
		return cacheData, nil
	}
	storageData, err := i.storage.ListUniq(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from cache")
	}
	return storageData, nil
}

func (i *IndicatorRepository) updateCache(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth int,
) {
	data, err := i.storage.List(ctx, exchange, symbol, unit, interval, name, depth, 500, 0)
	if err != nil {
		return
	}
	_ = i.cache.Save(ctx, data...)
}
