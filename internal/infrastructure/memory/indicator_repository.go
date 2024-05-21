package memory

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/metrics"
	"github.com/duke-git/lancet/v2/slice"
	lru "github.com/hashicorp/golang-lru/v2"
	"sync"
	"time"
)

var _ indicator.Repository = (*IndicatorRepository)(nil)

type IndicatorRepository struct {
	caches map[string]*lru.Cache[time.Time, indicator.StorageDTO]
	mu     sync.Mutex
	size   int
}

func NewIndicatorRepository(size int) indicator.Repository {
	caches := make(map[string]*lru.Cache[time.Time, indicator.StorageDTO])
	return &IndicatorRepository{caches: caches, size: size}
}

func (i *IndicatorRepository) Save(ctx context.Context, data ...indicator.StorageDTO) error {
	defer metrics.DBQueryHelper("memory", "indicators_save")()
	slice.SortBy[indicator.StorageDTO](data, func(a, b indicator.StorageDTO) bool {
		return a.Datetime.Before(b.Datetime)
	})
	i.mu.Lock()
	defer i.mu.Unlock()
	for _, item := range data {
		key := keyIndicatorFromDto(item)
		if i.caches[key] == nil {
			cache, err := lru.New[time.Time, indicator.StorageDTO](i.size)
			if err != nil {
				return err
			}
			i.caches[key] = cache
		}
		if i.caches[key].Contains(item.Datetime) {
			continue
		}
		i.caches[key].Add(item.Datetime, item)
	}

	return nil
}

func (i *IndicatorRepository) Find(
	ctx context.Context, exchange, symbol, unit string, interval int, datetime time.Time, name string, depth int,
) (*indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper("memory", "indicators_find")()
	i.mu.Lock()
	defer i.mu.Unlock()
	key := keyIndicator(exchange, symbol, unit, interval, name, depth)
	if _, has := i.caches[key]; !has {
		return nil, nil
	}
	if v, ok := i.caches[key].Get(datetime); ok {
		return &v, nil
	}

	return nil, nil
}

func (i *IndicatorRepository) List(ctx context.Context, exchange, symbol, unit string, interval int, name string, depth, limit, offset int) ([]indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper("memory", "indicators_list")()
	i.mu.Lock()
	defer i.mu.Unlock()
	key := keyIndicator(exchange, symbol, unit, interval, name, depth)
	if _, has := i.caches[key]; !has {
		return nil, nil
	}
	values := i.caches[key].Values()
	if len(values) <= offset {
		return nil, nil
	}
	slice.SortBy[indicator.StorageDTO](values, func(a, b indicator.StorageDTO) bool {
		return a.Datetime.After(b.Datetime)
	})
	if len(values) < limit+offset {
		return values[offset:], nil
	}
	return values[offset:limit], nil
}

func (i *IndicatorRepository) Last(ctx context.Context, exchange, symbol, unit string, interval int, name string, depth int) (*indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper("memory", "indicators_last")()
	i.mu.Lock()
	defer i.mu.Unlock()
	key := keyIndicator(exchange, symbol, unit, interval, name, depth)
	if _, has := i.caches[key]; !has {
		return nil, nil
	}
	values := i.caches[key].Values()
	if len(values) == 0 {
		return nil, nil
	}
	slice.SortBy[indicator.StorageDTO](values, func(a, b indicator.StorageDTO) bool {
		return a.Datetime.After(b.Datetime)
	})

	return &values[0], nil
}

func (i *IndicatorRepository) DeleteOldRows(
	ctx context.Context, symbol, exchangeName, unit string, interval int, name string, depth int, to time.Time,
) error {
	defer metrics.DBQueryHelper("memory", "indicators_delete_old_rows")()
	i.mu.Lock()
	defer i.mu.Unlock()
	key := keyIndicator(exchangeName, symbol, unit, interval, name, depth)
	if _, has := i.caches[key]; !has {
		return nil
	}
	values := i.caches[key].Values()
	i.caches[key].Purge()
	for _, value := range values {
		if value.Datetime.After(to) {
			i.caches[key].Add(value.Datetime, value)
		}
	}
	return nil
}

func (i *IndicatorRepository) ListUniq(ctx context.Context) ([]indicator.UniqDTO, error) {
	defer metrics.DBQueryHelper("memory", "indicators_list_uniq")()
	i.mu.Lock()
	defer i.mu.Unlock()
	uniqs := make([]indicator.UniqDTO, 0, 100)
	for _, cache := range i.caches {
		if _, v, ok := cache.GetOldest(); ok {
			uniqs = append(uniqs, indicator.UniqDTO{
				Symbol:   v.Symbol,
				Exchange: v.Exchange,
				Unit:     v.Unit,
				Interval: v.Interval,
				Name:     v.Name,
				Depth:    v.Depth,
			})
		}
	}
	return uniqs, nil
}

func keyIndicator(exchangeName, symbol, unit string, interval int, name string, depth int) string {
	return fmt.Sprintf("%s-%s-%s-%d-%s-%d", exchangeName, symbol, unit, interval, name, depth)
}

func keyIndicatorFromDto(data indicator.StorageDTO) string {
	return keyIndicator(data.Exchange, data.Symbol, data.Unit, data.Interval, data.Name, data.Depth)
}
