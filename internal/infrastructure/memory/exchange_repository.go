package memory

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/go-template/pkg/cache"
	"github.com/AlekseyPorandaykin/go-template/pkg/metrics"
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

func (e *ExchangeRepository) InfoByCategory(ctx context.Context, exchangeName, category string) ([]exchange.SymbolInfoStorageDTO, error) {
	defer metrics.DBQueryHelper("memory", "exchange_fetch_infos_by_category")()
	result := make([]exchange.SymbolInfoStorageDTO, 0, 1_000)
	for _, item := range e.data.ValuesByFirstKey(exchangeName) {
		if item.Category == category {
			result = append(result, item)
		}
	}
	return result, nil
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

func (e *ExchangeRepository) QuoteAssets(ctx context.Context) ([]string, error) {
	defer metrics.DBQueryHelper("memory", "exchange_quote_assets")()
	uniqQuoteAsset := make(map[string]struct{})
	quoteAssets := make([]string, 0, 100)
	for _, item := range e.data.Values() {
		if _, has := uniqQuoteAsset[item.QuoteAsset]; has {
			continue
		}
		quoteAssets = append(quoteAssets, item.QuoteAsset)
		uniqQuoteAsset[item.QuoteAsset] = struct{}{}
	}
	return quoteAssets, nil
}
