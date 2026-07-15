package repository

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/pkg/errors"
)

var _ candle_indicator.Repository = (*CandleIndicatorRepository)(nil)

type CandleIndicatorRepository struct {
	storage candle_indicator.Repository
	cache   candle_indicator.Repository
}

func NewCandleIndicatorRepository(storage, cache candle_indicator.Repository) candle_indicator.Repository {
	return &CandleIndicatorRepository{storage: storage, cache: cache}
}

func (repo *CandleIndicatorRepository) Save(ctx context.Context, data []candle_indicator.StorageDTO) error {
	if err := repo.cache.Save(ctx, data); err != nil {
		return errors.Wrap(err, "from cache")
	}
	if err := repo.storage.Save(ctx, data); err != nil {
		return errors.Wrap(err, "from storage")
	}
	return nil
}

func (repo *CandleIndicatorRepository) Find(ctx context.Context, name, exchange, symbol, unit string, interval int, startTime time.Time) (*candle_indicator.StorageDTO, error) {
	cacheData, errCache := repo.cache.Find(ctx, name, exchange, symbol, unit, interval, startTime)
	if errCache != nil {
		return nil, errors.Wrap(errCache, "from cache")
	}
	if cacheData != nil {
		return cacheData, nil
	}
	storageData, errStorage := repo.storage.Find(ctx, name, exchange, symbol, unit, interval, startTime)
	if errStorage != nil {
		return nil, errors.Wrap(errStorage, "from storage")
	}
	if storageData != nil {
		_ = repo.cache.Save(ctx, []candle_indicator.StorageDTO{*storageData})
	}
	return storageData, nil
}

func (repo *CandleIndicatorRepository) FindMany(
	ctx context.Context, name, exchange, symbol, unit string, interval int, startTimes []time.Time,
) ([]candle_indicator.StorageDTO, error) {
	if len(startTimes) == 0 {
		return nil, nil
	}
	cacheData, errCache := repo.cache.FindMany(ctx, name, exchange, symbol, unit, interval, startTimes)
	if errCache != nil {
		return nil, errors.Wrap(errCache, "from cache")
	}
	found := make(map[time.Time]struct{}, len(cacheData))
	for _, item := range cacheData {
		found[item.StartTime] = struct{}{}
	}
	missing := make([]time.Time, 0, len(startTimes)-len(cacheData))
	for _, startTime := range startTimes {
		if _, ok := found[startTime]; !ok {
			missing = append(missing, startTime)
		}
	}
	if len(missing) == 0 {
		return cacheData, nil
	}
	storageData, errStorage := repo.storage.FindMany(ctx, name, exchange, symbol, unit, interval, missing)
	if errStorage != nil {
		return nil, errors.Wrap(errStorage, "from storage")
	}
	if len(storageData) > 0 {
		_ = repo.cache.Save(ctx, storageData)
	}
	return append(cacheData, storageData...), nil
}

func (repo *CandleIndicatorRepository) FetchLast(ctx context.Context, name, exchange, symbol, unit string, interval int) ([]candle_indicator.StorageDTO, error) {
	cacheData, errCache := repo.cache.FetchLast(ctx, name, exchange, symbol, unit, interval)
	if errCache != nil {
		return nil, errors.Wrap(errCache, "from cache")
	}
	if len(cacheData) > 0 {
		return cacheData, nil
	}
	storageData, errStorage := repo.storage.FetchLast(ctx, name, exchange, symbol, unit, interval)
	if errStorage != nil {
		return nil, errors.Wrap(errStorage, "from storage")
	}
	_ = repo.cache.Save(ctx, storageData)
	return storageData, nil
}

func (repo *CandleIndicatorRepository) LastAddedFromDate(ctx context.Context, name, exchange, unit string, interval int, from time.Time) ([]candle_indicator.StorageDTO, error) {
	//Всегда берем из хранилища последние добавленные данные
	return repo.storage.LastAddedFromDate(ctx, name, exchange, unit, interval, from)
}
