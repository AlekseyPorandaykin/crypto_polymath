package repository

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/pkg/errors"
	"time"
)

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
