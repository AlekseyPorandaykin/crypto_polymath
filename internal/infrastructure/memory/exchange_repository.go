package memory

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/cache"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/metrics"
	"time"
)

type ExchangeRepository struct {
	data *cache.TwoQueue[string, string, exchange.SymbolInfoStorageDTO]
}

func NewExchangeRepository() *ExchangeRepository {
	return &ExchangeRepository{data: cache.NewTwoQueueCache[string, string, exchange.SymbolInfoStorageDTO]()}
}

func (e *ExchangeRepository) SaveSymbolInfo(ctx context.Context, data []exchange.SymbolInfoStorageDTO) error {
	defer metrics.DBQueryHelper("memory", "exchange_save_symbol_info")()
	for _, item := range data {
		e.data.Set(item.Exchange, item.Symbol, item)
	}
	return nil
}

func (e *ExchangeRepository) InfoBySymbol(ctx context.Context, exchangeName, symbol string) (*exchange.SymbolInfoStorageDTO, error) {
	defer metrics.DBQueryHelper("memory", "exchange_fetch_info_by_symbol")()
	return e.data.Get(exchangeName, symbol), nil
}

func (e *ExchangeRepository) DeleteOldRows(ctx context.Context, exchangeName string, to time.Time) error {
	defer metrics.DBQueryHelper("memory", "exchange_delete_old_rows")()
	e.data.DeleteByCondition(func(item exchange.SymbolInfoStorageDTO) bool {
		return item.Exchange == exchangeName && item.CreatedAt.Before(to)
	})
	return nil
}

func (e *ExchangeRepository) ListByExchange(ctx context.Context, exchangeName string) ([]exchange.SymbolInfoStorageDTO, error) {
	defer metrics.DBQueryHelper("memory", "exchange_list_by_exchange")()
	return e.data.ValuesByFirstKey(exchangeName), nil
}
