package logging

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/exchange"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/util"
	"go.uber.org/zap"
)

var _ exchange.Repository = (*ExchangeRepository)(nil)

type ExchangeRepository struct {
	inner  exchange.Repository
	logger *zap.Logger
	db     string
}

func NewExchangeRepository(inner exchange.Repository, logger *zap.Logger, db string) *ExchangeRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ExchangeRepository{inner: inner, logger: logger, db: db}
}

func (r *ExchangeRepository) SaveSymbolInfo(ctx context.Context, data []exchange.SymbolInfoStorageDTO) error {
	defer r.log(ctx, "exchange_save_symbol_info")()
	return r.inner.SaveSymbolInfo(ctx, data)
}

func (r *ExchangeRepository) InfoBySymbol(ctx context.Context, exchangeName, symbol string) (*exchange.SymbolInfoStorageDTO, error) {
	defer r.log(ctx, "exchange_fetch_info_by_symbol")()
	return r.inner.InfoBySymbol(ctx, exchangeName, symbol)
}

func (r *ExchangeRepository) InfoByCategory(ctx context.Context, exchangeName, category string) ([]exchange.SymbolInfoStorageDTO, error) {
	defer r.log(ctx, "exchange_fetch_infos_by_category")()
	return r.inner.InfoByCategory(ctx, exchangeName, category)
}

func (r *ExchangeRepository) QuoteAssets(ctx context.Context) ([]string, error) {
	defer r.log(ctx, "exchange_quote_assets")()
	return r.inner.QuoteAssets(ctx)
}

func (r *ExchangeRepository) DeleteOldRows(ctx context.Context, exchangeName string, to time.Time) error {
	defer r.log(ctx, "exchange_delete_old_rows")()
	return r.inner.DeleteOldRows(ctx, exchangeName, to)
}

func (r *ExchangeRepository) ListByExchange(ctx context.Context, exchangeName string) ([]exchange.SymbolInfoStorageDTO, error) {
	defer r.log(ctx, "list_by_exchange")()
	return r.inner.ListByExchange(ctx, exchangeName)
}

func (r *ExchangeRepository) log(ctx context.Context, query string) func() {
	now := time.Now()
	return func() {
		r.logger.Debug("Execute query",
			zap.String("query", query),
			zap.String("db", r.db),
			zap.String("duration", time.Since(now).String()),
			zap.String("request_id", util.RequestIDFromContext(ctx)),
		)
	}
}
