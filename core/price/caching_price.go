package price

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type cachingPrice struct {
	cache   Price
	service Price
}

func NewCachingPrice(cache, service Price) Price {
	return &cachingPrice{cache: cache, service: service}
}

func (a *cachingPrice) AddLoader(exchange string, loader ExchangeLoader) {
	a.service.AddLoader(exchange, loader)
	a.cache.AddLoader(exchange, loader)
}

func (a *cachingPrice) LoadPrices(ctx context.Context, exchange string) ([]domain.Price, error) {
	cacheData, errCache := a.cache.LoadPrices(ctx, exchange)
	if errCache != nil {
		zap.L().Error("load prices from cache", zap.Error(errCache))
	}
	if cacheData != nil {
		return cacheData, nil
	}
	data, err := a.service.LoadPrices(ctx, exchange)
	if err != nil {
		return nil, err
	}
	if err := a.cache.Save(ctx, data...); err != nil {
		return nil, errors.Wrap(err, "save price to cache")
	}
	return data, nil
}

func (a *cachingPrice) LastPrice(ctx context.Context, exchange, symbol string) (*domain.Price, error) {
	cacheData, errCache := a.cache.LastPrice(ctx, exchange, symbol)
	if errCache != nil {
		zap.L().Error("get last price from cache", zap.Error(errCache))
	}
	if cacheData != nil {
		return cacheData, nil
	}
	data, err := a.service.LastPrice(ctx, exchange, symbol)
	if err != nil {
		return nil, err
	}
	if data != nil {
		a.updateCache(ctx, *data)
	}
	return data, nil
}

func (a *cachingPrice) LastPricesByExchange(ctx context.Context, exchange string) ([]domain.Price, error) {
	cacheData, errCache := a.cache.LastPricesByExchange(ctx, exchange)
	if errCache != nil {
		zap.L().Error("get last prices from cache", zap.Error(errCache))
	}
	if cacheData != nil {
		return cacheData, nil
	}
	data, err := a.service.LastPricesByExchange(ctx, exchange)
	if err != nil {
		return nil, err
	}
	a.updateCache(ctx, data...)
	return data, nil
}

func (a *cachingPrice) LastPricesBySymbol(ctx context.Context, symbol string) ([]domain.Price, error) {
	cacheData, errCache := a.cache.LastPricesBySymbol(ctx, symbol)
	if errCache != nil {
		zap.L().Error("get last prices from cache", zap.Error(errCache))
	}
	if cacheData != nil {
		return cacheData, nil
	}
	data, err := a.service.LastPricesBySymbol(ctx, symbol)
	if err != nil {
		return nil, err
	}
	a.updateCache(ctx, data...)
	return data, nil
}

func (a *cachingPrice) Save(ctx context.Context, data ...domain.Price) error {
	if err := a.service.Save(ctx, data...); err != nil {
		return err
	}
	if err := a.cache.Save(ctx, data...); err != nil {
		return err
	}

	return nil
}

func (a *cachingPrice) DeleteOldRaws(ctx context.Context, exchange string, to time.Time) error {
	return a.service.DeleteOldRaws(ctx, exchange, to)
}

func (a *cachingPrice) updateCache(ctx context.Context, data ...domain.Price) {
	_ = a.cache.Save(ctx, data...)
}
