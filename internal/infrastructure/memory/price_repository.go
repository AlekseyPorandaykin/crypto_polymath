package memory

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/go-template/pkg/cache"
	"github.com/AlekseyPorandaykin/go-template/pkg/metrics"
	"time"
)

var _ price.Repository = (*PriceRepository)(nil)

type PriceRepository struct {
	data *cache.TwoQueue[string, string, price.StorageDTO]
}

func NewPriceRepository() *PriceRepository {
	return &PriceRepository{
		data: cache.NewTwoQueueCache[string, string, price.StorageDTO](),
	}
}

func (c *PriceRepository) Save(ctx context.Context, data ...price.StorageDTO) error {
	defer metrics.CacheQueryHelper("memory", "price_save")()
	for _, item := range data {
		c.data.Set(item.Exchange, item.Symbol, item)
	}
	return nil
}

func (c *PriceRepository) Find(ctx context.Context, exchange, symbol string) (*price.StorageDTO, error) {
	defer metrics.CacheQueryHelper("memory", "price_find")()
	return c.data.Get(exchange, symbol), nil
}

func (c *PriceRepository) ListByExchange(ctx context.Context, exchange string) ([]price.StorageDTO, error) {
	defer metrics.CacheQueryHelper("memory", "price_list_by_exchange")()
	return c.data.ValuesByFirstKey(exchange), nil
}

func (c *PriceRepository) ListBySymbol(ctx context.Context, symbol string) ([]price.StorageDTO, error) {
	defer metrics.CacheQueryHelper("memory", "price_list_by_symbol")()
	return c.data.ValuesBySecondKey(symbol), nil
}

func (c *PriceRepository) Delete(ctx context.Context, exchange string, to time.Time) error {
	defer metrics.CacheQueryHelper("memory", "price_delete")()
	c.data.DeleteByCondition(func(item price.StorageDTO) bool {
		return item.Exchange == exchange && item.CreatedAt.Before(to)
	})
	return nil
}
