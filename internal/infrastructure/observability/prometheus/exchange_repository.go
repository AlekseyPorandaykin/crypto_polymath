package prometheus

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/go-kit/pkg/metrics"
)

var _ exchange.Repository = (*ExchangeRepository)(nil)

type ExchangeRepository struct {
	inner exchange.Repository
	db    string
}

func NewExchangeRepository(inner exchange.Repository, db string) *ExchangeRepository {
	return &ExchangeRepository{inner: inner, db: db}
}

func (r *ExchangeRepository) SaveSymbolInfo(ctx context.Context, data []exchange.SymbolInfoStorageDTO) error {
	defer metrics.DBQueryHelper(r.db, "exchange_save_symbol_info")()
	return r.inner.SaveSymbolInfo(ctx, data)
}

func (r *ExchangeRepository) InfoBySymbol(ctx context.Context, exchangeName, symbol string) (*exchange.SymbolInfoStorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "exchange_fetch_info_by_symbol")()
	return r.inner.InfoBySymbol(ctx, exchangeName, symbol)
}

func (r *ExchangeRepository) InfoByCategory(ctx context.Context, exchangeName, category string) ([]exchange.SymbolInfoStorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "exchange_fetch_infos_by_category")()
	return r.inner.InfoByCategory(ctx, exchangeName, category)
}

func (r *ExchangeRepository) QuoteAssets(ctx context.Context) ([]string, error) {
	defer metrics.DBQueryHelper(r.db, "exchange_quote_assets")()
	return r.inner.QuoteAssets(ctx)
}

func (r *ExchangeRepository) DeleteOldRows(ctx context.Context, exchangeName string, to time.Time) error {
	defer metrics.DBQueryHelper(r.db, "exchange_delete_old_rows")()
	return r.inner.DeleteOldRows(ctx, exchangeName, to)
}

func (r *ExchangeRepository) ListByExchange(ctx context.Context, exchangeName string) ([]exchange.SymbolInfoStorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "list_by_exchange")()
	return r.inner.ListByExchange(ctx, exchangeName)
}
