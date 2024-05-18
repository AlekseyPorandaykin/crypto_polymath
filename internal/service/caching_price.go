package service

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/cache"
	"time"
)

type CachingPrice struct {
	data *cache.TwoQueueCache[string, string, domain.Price]
}

func NewCachingPrice() *CachingPrice {
	return &CachingPrice{
		data: cache.NewTwoQueueCache[string, string, domain.Price](),
	}
}

func (c *CachingPrice) AddLoader(exchange string, loader price.ExchangeLoader) {
	return
}

func (c *CachingPrice) LoadPrices(ctx context.Context, exchange string) ([]domain.Price, error) {
	return nil, nil
}

func (c *CachingPrice) DeleteOldRaws(ctx context.Context, exchange string, to time.Time) error {
	return nil
}

func (c *CachingPrice) Save(ctx context.Context, data ...domain.Price) error {
	for _, item := range data {
		c.data.Set(item.Exchange, item.Symbol, item)
	}
	return nil
}

func (c *CachingPrice) LastPrice(ctx context.Context, exchange, symbol string) (*domain.Price, error) {
	return c.data.Get(exchange, symbol), nil
}

func (c *CachingPrice) LastPricesByExchange(ctx context.Context, exchange string) ([]domain.Price, error) {
	return c.data.SecondQueueValues(exchange), nil
}

func (c *CachingPrice) LastPricesBySymbol(ctx context.Context, symbol string) ([]domain.Price, error) {
	values := c.data.Values()
	prices := make([]domain.Price, 0, 100)
	for _, value := range values {
		if value.Symbol != symbol {
			continue
		}
		prices = append(prices, value)
	}
	return prices, nil
}
