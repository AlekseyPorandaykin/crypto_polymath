package prometheus

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/go-template/pkg/metrics"
)

var _ candlestick.Repository = (*CandlestickRepository)(nil)

type CandlestickRepository struct {
	inner candlestick.Repository
	db    string
}

func NewCandlestickRepository(inner candlestick.Repository, db string) *CandlestickRepository {
	return &CandlestickRepository{inner: inner, db: db}
}

func (r *CandlestickRepository) Save(ctx context.Context, data ...candlestick.StorageDTO) error {
	defer metrics.DBQueryHelper(r.db, "candlestick_save")()
	return r.inner.Save(ctx, data...)
}

func (r *CandlestickRepository) Last(ctx context.Context, exchange, symbol, unit string, interval, limit, offset int) ([]candlestick.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "candlestick_last")()
	return r.inner.Last(ctx, exchange, symbol, unit, interval, limit, offset)
}

func (r *CandlestickRepository) LastToDate(ctx context.Context, exchange, symbol, unit string, interval, limit int, to time.Time) ([]candlestick.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "candlestick_last_to_date")()
	return r.inner.LastToDate(ctx, exchange, symbol, unit, interval, limit, to)
}

func (r *CandlestickRepository) FromDate(ctx context.Context, exchange, symbol, unit string, interval, limit int, to time.Time) ([]candlestick.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "candlestick_find_from")()
	return r.inner.FromDate(ctx, exchange, symbol, unit, interval, limit, to)
}

func (r *CandlestickRepository) DeleteOldRows(ctx context.Context, exchange, symbol, unit string, interval int, to time.Time) error {
	defer metrics.DBQueryHelper(r.db, "candlestick_delete_old_rows")()
	return r.inner.DeleteOldRows(ctx, exchange, symbol, unit, interval, to)
}

func (r *CandlestickRepository) DeletePrevRows(ctx context.Context, exchange, symbol, unit string, interval int, to time.Time) error {
	defer metrics.DBQueryHelper(r.db, "candlestick_delete_prev_rows")()
	return r.inner.DeletePrevRows(ctx, exchange, symbol, unit, interval, to)
}

func (r *CandlestickRepository) ListUniq(ctx context.Context) ([]candlestick.UniqDTO, error) {
	defer metrics.DBQueryHelper(r.db, "list_uniq")()
	return r.inner.ListUniq(ctx)
}

func (r *CandlestickRepository) AllSymbols(ctx context.Context) ([]string, error) {
	defer metrics.DBQueryHelper(r.db, "all_symbols")()
	return r.inner.AllSymbols(ctx)
}
