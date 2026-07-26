package memory

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/go-kit/pkg/metrics"
	"github.com/duke-git/lancet/v2/slice"
	lru "github.com/hashicorp/golang-lru/v2"
	"sync"
	"time"
)

var _ candlestick.Repository = (*CandlestickRepository)(nil)

type CandlestickRepository struct {
	caches map[string]*lru.Cache[time.Time, candlestick.StorageDTO]
	mu     sync.Mutex

	size int
}

func NewCandlestickRepository(size int) candlestick.Repository {
	caches := make(map[string]*lru.Cache[time.Time, candlestick.StorageDTO])
	return &CandlestickRepository{caches: caches, size: size}
}

func (c *CandlestickRepository) Save(ctx context.Context, data ...candlestick.StorageDTO) error {
	defer metrics.CacheQueryHelper("memory", "candlestick_save")()
	slice.SortBy[candlestick.StorageDTO](data, func(a, b candlestick.StorageDTO) bool {
		return a.StartTime.Before(b.StartTime)
	})
	c.mu.Lock()
	defer c.mu.Unlock()
	existKey := make(map[string]map[time.Time]bool, len(data))
	for _, item := range data {
		key := keyCandlestickFromDto(item)
		if _, has := existKey[key]; !has {
			existKey[key] = make(map[time.Time]bool)
		}
		if c.caches[key] == nil {
			cache, err := lru.New[time.Time, candlestick.StorageDTO](c.size)
			if err != nil {
				return err
			}
			c.caches[key] = cache
		}
		if existKey[key][item.StartTime] {
			continue
		}
		c.caches[key].Add(item.StartTime, item)
		existKey[key][item.StartTime] = true
	}
	return nil
}

func (c *CandlestickRepository) Last(ctx context.Context, exchange, symbol, unit string, interval, limit, offset int) ([]candlestick.StorageDTO, error) {
	defer metrics.CacheQueryHelper("memory", "candlestick_last")()
	c.mu.Lock()
	defer c.mu.Unlock()
	cache := c.caches[createKeyCandlestick(exchange, symbol, unit, interval)]
	if cache == nil {
		return nil, nil
	}
	if offset >= cache.Len() {
		return nil, nil
	}
	data := cache.Values()
	slice.SortBy[candlestick.StorageDTO](data, func(a, b candlestick.StorageDTO) bool {
		return a.StartTime.After(b.StartTime)
	})
	if len(data) < offset+limit {
		return data[offset:], nil
	}

	return data[offset : offset+limit], nil
}

func (c *CandlestickRepository) DeleteOldRows(ctx context.Context, exchange, symbol, unit string, interval int, to time.Time) error {
	defer metrics.CacheQueryHelper("memory", "candlestick_delete_old_rows")()
	c.mu.Lock()
	defer c.mu.Unlock()
	cache := c.caches[createKeyCandlestick(exchange, symbol, unit, interval)]
	if cache == nil {
		return nil
	}
	for _, datetime := range cache.Keys() {
		if datetime.Before(to) {
			cache.Remove(datetime)
		}
	}

	return nil
}
func (c *CandlestickRepository) DeletePrevRows(ctx context.Context, exchange, symbol, unit string, interval int, to time.Time) error {
	defer metrics.CacheQueryHelper("memory", "candlestick_delete_prev_rows")()
	c.mu.Lock()
	defer c.mu.Unlock()
	cache := c.caches[createKeyCandlestick(exchange, symbol, unit, interval)]
	if cache == nil {
		return nil
	}

	for _, v := range cache.Values() {
		if v.CreatedAt.Before(to) {
			cache.Remove(v.StartTime)
		}
	}

	return nil
}

func (c *CandlestickRepository) LastToDate(ctx context.Context, exchange, symbol, unit string, interval, limit int, to time.Time) ([]candlestick.StorageDTO, error) {
	defer metrics.CacheQueryHelper("memory", "candlestick_last_to_date")()
	c.mu.Lock()
	defer c.mu.Unlock()
	cache := c.caches[createKeyCandlestick(exchange, symbol, unit, interval)]
	if cache == nil {
		return nil, nil
	}
	data := make([]candlestick.StorageDTO, 0, limit)
	datetimes := cache.Keys()
	for _, datetime := range datetimes {
		if !datetime.After(to) {
			if v, ok := cache.Get(datetime); ok {
				data = append(data, v)
			}
		}
	}
	if len(data) == 0 {
		return nil, nil
	}
	slice.SortBy[candlestick.StorageDTO](data, func(a, b candlestick.StorageDTO) bool {
		return a.StartTime.After(b.StartTime)
	})
	if len(data) > limit {
		return data[:limit], nil
	}
	return data, nil
}

func (c *CandlestickRepository) FromDate(ctx context.Context, exchange, symbol, unit string, interval, limit int, to time.Time) ([]candlestick.StorageDTO, error) {
	defer metrics.CacheQueryHelper("memory", "candlestick_find_from")()
	c.mu.Lock()
	defer c.mu.Unlock()
	cache := c.caches[createKeyCandlestick(exchange, symbol, unit, interval)]
	if cache == nil {
		return nil, nil
	}
	data := make([]candlestick.StorageDTO, 0, limit)
	datetimes := cache.Keys()
	for _, datetime := range datetimes {
		if !datetime.Before(to) {
			if v, ok := cache.Get(datetime); ok {
				data = append(data, v)
			}
		}
	}
	if len(data) == 0 {
		return nil, nil
	}
	slice.SortBy[candlestick.StorageDTO](data, func(a, b candlestick.StorageDTO) bool {
		return a.StartTime.Before(b.StartTime)
	})
	if len(data) > limit {
		return data[:limit], nil
	}

	return data, nil
}

func (c *CandlestickRepository) ListUniq(ctx context.Context) ([]candlestick.UniqDTO, error) {
	defer metrics.CacheQueryHelper("memory", "indicators_list_uniq")()
	c.mu.Lock()
	defer c.mu.Unlock()
	data := make([]candlestick.UniqDTO, 0, len(c.caches))
	for _, cache := range c.caches {
		if _, v, ok := cache.GetOldest(); ok {
			data = append(data, candlestick.UniqDTO{
				Symbol:   v.Symbol,
				Exchange: v.Exchange,
				Unit:     v.Unit,
				Interval: v.Interval,
			})
		}
	}
	return data, nil
}

func keyCandlestickFromDto(data candlestick.StorageDTO) string {
	return createKeyCandlestick(data.Exchange, data.Symbol, data.Unit, data.Interval)
}

func createKeyCandlestick(exchangeName, symbol, unit string, interval int) string {
	return fmt.Sprintf("%s-%s-%s-%d", exchangeName, symbol, unit, interval)
}

func (c *CandlestickRepository) AllSymbols(ctx context.Context) ([]string, error) {
	return nil, nil
}
