package prometheus

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/go-kit/pkg/metrics"
)

var _ candle_indicator.Repository = (*CandleIndicatorRepository)(nil)

type CandleIndicatorRepository struct {
	inner candle_indicator.Repository
	db    string
}

func NewCandleIndicatorRepository(inner candle_indicator.Repository, db string) *CandleIndicatorRepository {
	return &CandleIndicatorRepository{inner: inner, db: db}
}

func (r *CandleIndicatorRepository) Save(ctx context.Context, data []candle_indicator.StorageDTO) error {
	defer metrics.DBQueryHelper(r.db, "candle_indicator_save")()
	return r.inner.Save(ctx, data)
}

func (r *CandleIndicatorRepository) Find(
	ctx context.Context, name, exchange, symbol, unit string, interval int, from time.Time,
) (*candle_indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "candle_indicator_find")()
	return r.inner.Find(ctx, name, exchange, symbol, unit, interval, from)
}

func (r *CandleIndicatorRepository) FindMany(
	ctx context.Context, name, exchange, symbol, unit string, interval int, startTimes []time.Time,
) ([]candle_indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "candle_indicator_find_many")()
	return r.inner.FindMany(ctx, name, exchange, symbol, unit, interval, startTimes)
}

func (r *CandleIndicatorRepository) FetchLast(
	ctx context.Context, name, exchange, symbol, unit string, interval int,
) ([]candle_indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "candle_indicator_fetch_last")()
	return r.inner.FetchLast(ctx, name, exchange, symbol, unit, interval)
}

func (r *CandleIndicatorRepository) LastAddedFromDate(
	ctx context.Context, name, exchange, unit string, interval int, from time.Time,
) ([]candle_indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "candle_indicator_fetch_last")()
	return r.inner.LastAddedFromDate(ctx, name, exchange, unit, interval, from)
}
