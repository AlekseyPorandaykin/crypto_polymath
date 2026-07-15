package repository

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/pkg/errors"
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
	_ = i.cache.Save(ctx, data...)
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
	if storageData == nil {
		return nil, nil
	}
	i.updateCache(ctx, exchange, symbol, unit, interval, name, depth)
	return storageData, nil
}

func (i *IndicatorRepository) FindMany(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth int, datetimes []time.Time,
) ([]indicator.StorageDTO, error) {
	if len(datetimes) == 0 {
		return nil, nil
	}
	cacheData, err := i.cache.FindMany(ctx, exchange, symbol, unit, interval, name, depth, datetimes)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from cache")
	}
	found := make(map[time.Time]struct{}, len(cacheData))
	for _, item := range cacheData {
		found[item.Datetime] = struct{}{}
	}
	missing := make([]time.Time, 0, len(datetimes)-len(cacheData))
	for _, datetime := range datetimes {
		if _, ok := found[datetime]; !ok {
			missing = append(missing, datetime)
		}
	}
	if len(missing) == 0 {
		return cacheData, nil
	}
	storageData, err := i.storage.FindMany(ctx, exchange, symbol, unit, interval, name, depth, missing)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from storage")
	}
	if len(storageData) > 0 {
		_ = i.cache.Save(ctx, storageData...)
	}
	return append(cacheData, storageData...), nil
}

func (i *IndicatorRepository) FindManyByName(
	ctx context.Context, exchange, symbol, unit string, interval int, datetime time.Time, depth int, names []string,
) ([]indicator.StorageDTO, error) {
	if len(names) == 0 {
		return nil, nil
	}
	cacheData, err := i.cache.FindManyByName(ctx, exchange, symbol, unit, interval, datetime, depth, names)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from cache")
	}
	found := make(map[string]struct{}, len(cacheData))
	for _, item := range cacheData {
		found[item.Name] = struct{}{}
	}
	missing := make([]string, 0, len(names)-len(cacheData))
	for _, name := range names {
		if _, ok := found[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) == 0 {
		return cacheData, nil
	}
	storageData, err := i.storage.FindManyByName(ctx, exchange, symbol, unit, interval, datetime, depth, missing)
	if err != nil {
		return nil, errors.Wrap(err, "fetch from storage")
	}
	if len(storageData) > 0 {
		_ = i.cache.Save(ctx, storageData...)
	}
	return append(cacheData, storageData...), nil
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
	if len(storageData) == 0 {
		return nil, nil
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
	if storageData == nil {
		return nil, nil
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

func (i *IndicatorRepository) LastToDate(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth, limit int, to time.Time,
) ([]indicator.StorageDTO, error) {
	return i.storage.LastToDate(ctx, exchange, symbol, unit, interval, name, depth, limit, to)
}

func (i *IndicatorRepository) updateCache(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth int,
) {
	data, err := i.storage.List(ctx, exchange, symbol, unit, interval, name, depth, 100, 0)
	if err != nil {
		return
	}
	_ = i.cache.Save(ctx, data...)
}

func (i *IndicatorRepository) AllIndicatorInfo(ctx context.Context) (map[string][]indicator.IndicatorInfo, error) {
	return i.storage.AllIndicatorInfo(ctx)
}
