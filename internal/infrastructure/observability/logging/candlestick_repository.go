package logging

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/util"
	"go.uber.org/zap"
)

var _ candlestick.Repository = (*CandlestickRepository)(nil)

type CandlestickRepository struct {
	inner  candlestick.Repository
	logger *zap.Logger
	db     string
}

func NewCandlestickRepository(inner candlestick.Repository, logger *zap.Logger, db string) *CandlestickRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &CandlestickRepository{inner: inner, logger: logger, db: db}
}

func (r *CandlestickRepository) Save(ctx context.Context, data ...candlestick.StorageDTO) error {
	defer r.log(ctx, "candlestick_save")()
	return r.inner.Save(ctx, data...)
}

func (r *CandlestickRepository) Last(ctx context.Context, exchange, symbol, unit string, interval, limit, offset int) ([]candlestick.StorageDTO, error) {
	defer r.log(ctx, "candlestick_last")()
	return r.inner.Last(ctx, exchange, symbol, unit, interval, limit, offset)
}

func (r *CandlestickRepository) LastToDate(ctx context.Context, exchange, symbol, unit string, interval, limit int, to time.Time) ([]candlestick.StorageDTO, error) {
	defer r.log(ctx, "candlestick_last_to_date")()
	return r.inner.LastToDate(ctx, exchange, symbol, unit, interval, limit, to)
}

func (r *CandlestickRepository) FromDate(ctx context.Context, exchange, symbol, unit string, interval, limit int, to time.Time) ([]candlestick.StorageDTO, error) {
	defer r.log(ctx, "candlestick_find_from")()
	return r.inner.FromDate(ctx, exchange, symbol, unit, interval, limit, to)
}

func (r *CandlestickRepository) DeleteOldRows(ctx context.Context, exchange, symbol, unit string, interval int, to time.Time) error {
	defer r.log(ctx, "candlestick_delete_old_rows")()
	return r.inner.DeleteOldRows(ctx, exchange, symbol, unit, interval, to)
}

func (r *CandlestickRepository) DeletePrevRows(ctx context.Context, exchange, symbol, unit string, interval int, to time.Time) error {
	defer r.log(ctx, "candlestick_delete_prev_rows")()
	return r.inner.DeletePrevRows(ctx, exchange, symbol, unit, interval, to)
}

func (r *CandlestickRepository) ListUniq(ctx context.Context) ([]candlestick.UniqDTO, error) {
	defer r.log(ctx, "list_uniq")()
	return r.inner.ListUniq(ctx)
}

func (r *CandlestickRepository) AllSymbols(ctx context.Context) ([]string, error) {
	defer r.log(ctx, "all_symbols")()
	return r.inner.AllSymbols(ctx)
}

func (r *CandlestickRepository) log(ctx context.Context, query string) func() {
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
