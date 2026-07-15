package prometheus

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/go-template/pkg/metrics"
)

var _ indicator.Repository = (*IndicatorRepository)(nil)

type IndicatorRepository struct {
	inner indicator.Repository
	db    string
}

func NewIndicatorRepository(inner indicator.Repository, db string) *IndicatorRepository {
	return &IndicatorRepository{inner: inner, db: db}
}

func (r *IndicatorRepository) Save(ctx context.Context, data ...indicator.StorageDTO) error {
	defer metrics.DBQueryHelper(r.db, "indicators_save")()
	return r.inner.Save(ctx, data...)
}

func (r *IndicatorRepository) Find(
	ctx context.Context, exchange, symbol, unit string, interval int, datetime time.Time, name string, depth int,
) (*indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "indicators_find")()
	return r.inner.Find(ctx, exchange, symbol, unit, interval, datetime, name, depth)
}

func (r *IndicatorRepository) FindMany(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth int, datetimes []time.Time,
) ([]indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "indicators_find_many")()
	return r.inner.FindMany(ctx, exchange, symbol, unit, interval, name, depth, datetimes)
}

func (r *IndicatorRepository) FindManyByName(
	ctx context.Context, exchange, symbol, unit string, interval int, datetime time.Time, depth int, names []string,
) ([]indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "indicators_find_many")()
	return r.inner.FindManyByName(ctx, exchange, symbol, unit, interval, datetime, depth, names)
}

func (r *IndicatorRepository) List(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth, limit, offset int,
) ([]indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "indicators_list")()
	return r.inner.List(ctx, exchange, symbol, unit, interval, name, depth, limit, offset)
}

func (r *IndicatorRepository) Last(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth int,
) (*indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "indicators_last")()
	return r.inner.Last(ctx, exchange, symbol, unit, interval, name, depth)
}

func (r *IndicatorRepository) DeleteOldRows(
	ctx context.Context, symbol, exchangeName, unit string, interval int, name string, depth int, to time.Time,
) error {
	defer metrics.DBQueryHelper(r.db, "indicators_delete_old_rows")()
	return r.inner.DeleteOldRows(ctx, symbol, exchangeName, unit, interval, name, depth, to)
}

func (r *IndicatorRepository) ListUniq(ctx context.Context) ([]indicator.UniqDTO, error) {
	defer metrics.DBQueryHelper(r.db, "indicators_list_uniq")()
	return r.inner.ListUniq(ctx)
}

func (r *IndicatorRepository) LastToDate(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth, limit int, to time.Time,
) ([]indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "indicators_list")()
	return r.inner.LastToDate(ctx, exchange, symbol, unit, interval, name, depth, limit, to)
}

func (r *IndicatorRepository) AllIndicatorInfo(ctx context.Context) (map[string][]indicator.IndicatorInfo, error) {
	defer metrics.DBQueryHelper(r.db, "indicators_all_indicator_info_model")()
	return r.inner.AllIndicatorInfo(ctx)
}
