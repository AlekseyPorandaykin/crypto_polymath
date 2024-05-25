package adapters

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/pkg/errors"
	"time"
)

type ExchangeRepository struct {
	storage exchange.Repository
	cache   exchange.Repository
}

func NewExchangeRepository(storage, cache exchange.Repository) exchange.Repository {
	return &ExchangeRepository{storage: storage, cache: cache}
}

func (e *ExchangeRepository) SaveSymbolInfo(ctx context.Context, data []exchange.SymbolInfoStorageDTO) error {
	_ = e.cache.SaveSymbolInfo(ctx, data)
	if err := e.storage.SaveSymbolInfo(ctx, data); err != nil {
		return errors.Wrap(err, "save to storage")
	}
	return nil
}

func (e *ExchangeRepository) InfoBySymbol(ctx context.Context, exchangeName, symbol string) (*exchange.SymbolInfoStorageDTO, error) {
	dataCache, _ := e.cache.InfoBySymbol(ctx, exchangeName, symbol)
	if dataCache != nil {
		return dataCache, nil
	}
	dataStorage, err := e.storage.InfoBySymbol(ctx, exchangeName, symbol)
	if err != nil {
		return nil, errors.Wrap(err, "fetch data from storage")
	}
	e.updateCache(ctx, exchangeName)
	return dataStorage, nil
}

func (e *ExchangeRepository) DeleteOldRows(ctx context.Context, exchangeName string, to time.Time) error {
	_ = e.cache.DeleteOldRows(ctx, exchangeName, to)
	return e.storage.DeleteOldRows(ctx, exchangeName, to)
}

func (e *ExchangeRepository) ListByExchange(ctx context.Context, exchangeName string) ([]exchange.SymbolInfoStorageDTO, error) {
	dataCache, _ := e.cache.ListByExchange(ctx, exchangeName)
	if len(dataCache) > 0 {
		return dataCache, nil
	}
	dataStorage, err := e.storage.ListByExchange(ctx, exchangeName)
	if err != nil {
		return nil, err
	}
	e.updateCache(ctx, exchangeName)

	return dataStorage, nil
}

func (e *ExchangeRepository) updateCache(ctx context.Context, exchangeName string) {
	data, _ := e.storage.ListByExchange(ctx, exchangeName)
	_ = e.cache.SaveSymbolInfo(ctx, data)
}
