package memory

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/duke-git/lancet/v2/slice"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pkg/errors"
	"sync"
	"time"
)

var _ candle_indicator.Repository = (*CandleIndicatorRepository)(nil)

type CandleIndicatorRepository struct {
	caches map[string]*lru.Cache[time.Time, candle_indicator.StorageDTO]
	mu     sync.Mutex
	size   int
}

func NewCandleIndicatorRepository(size int) *CandleIndicatorRepository {
	return &CandleIndicatorRepository{
		caches: make(map[string]*lru.Cache[time.Time, candle_indicator.StorageDTO]),
		size:   size,
	}
}

func (repo *CandleIndicatorRepository) Save(ctx context.Context, data []candle_indicator.StorageDTO) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	for _, item := range data {
		key := createCandleIndicatorKey(item.Name, item.Exchange, item.Symbol, item.Unit, item.Interval)
		if _, has := repo.caches[key]; !has {
			cache, err := lru.New[time.Time, candle_indicator.StorageDTO](repo.size)
			if err != nil {
				return errors.Wrap(err, "failed to create cache")
			}
			repo.caches[key] = cache
		}
		repo.caches[key].Add(item.StartTime, item)
	}
	return nil
}

func (repo *CandleIndicatorRepository) Find(
	ctx context.Context, name, exchange, symbol, unit string, interval int, startTime time.Time,
) (*candle_indicator.StorageDTO, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	key := createCandleIndicatorKey(name, exchange, symbol, unit, interval)
	if _, has := repo.caches[key]; !has {
		cache, err := lru.New[time.Time, candle_indicator.StorageDTO](repo.size)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create cache")
		}
		repo.caches[key] = cache
		return nil, nil
	}
	res, ok := repo.caches[key].Get(startTime)
	if !ok {
		return nil, nil
	}
	return &res, nil
}

func (repo *CandleIndicatorRepository) FetchLast(ctx context.Context, name, exchange, symbol, unit string, interval int) ([]candle_indicator.StorageDTO, error) {
	data := make([]candle_indicator.StorageDTO, 0, 100)
	for _, cache := range repo.caches {
		for _, v := range cache.Values() {
			if v.Name == name && v.Exchange == exchange && v.Symbol == symbol && v.Unit == unit && v.Interval == interval {
				data = append(data, v)
			}
		}
	}
	slice.SortBy[candle_indicator.StorageDTO](data, func(a, b candle_indicator.StorageDTO) bool {
		return a.StartTime.Before(b.StartTime)
	})
	if len(data) > 100 {
		return data[:100], nil
	}
	return data, nil
}

func (repo *CandleIndicatorRepository) LastAddedFromDate(
	ctx context.Context, name, exchange, unit string, interval int, from time.Time,
) ([]candle_indicator.StorageDTO, error) {
	return nil, nil
}

func createCandleIndicatorKey(name, exchange, symbol, unit string, interval int) string {
	return fmt.Sprintf("%s-%s-%s-%s-%d", name, exchange, symbol, unit, interval)
}
