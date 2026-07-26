package prometheus

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/go-kit/pkg/metrics"
)

var _ analysis.Repository = (*AnalysisRepository)(nil)

type AnalysisRepository struct {
	inner analysis.Repository
	db    string
}

func NewAnalysisRepository(inner analysis.Repository, db string) *AnalysisRepository {
	return &AnalysisRepository{inner: inner, db: db}
}

func (r *AnalysisRepository) Save(ctx context.Context, data ...analysis.Analytic) error {
	defer metrics.DBQueryHelper(r.db, "analysis_save")()
	return r.inner.Save(ctx, data...)
}

func (r *AnalysisRepository) Find(
	ctx context.Context,
	name, exchangeName, symbol string,
	unit domain.Unit,
	datetime time.Time,
	interval, indicatorDepth, depth int,
) (*analysis.Analytic, error) {
	defer metrics.DBQueryHelper(r.db, "analysis_find")()
	return r.inner.Find(ctx, name, exchangeName, symbol, unit, datetime, interval, indicatorDepth, depth)
}

func (r *AnalysisRepository) FindMany(
	ctx context.Context,
	name, exchangeName, symbol string,
	unit domain.Unit,
	interval, indicatorDepth, depth int,
	datetimes []time.Time,
) ([]analysis.Analytic, error) {
	defer metrics.DBQueryHelper(r.db, "analysis_find_many")()
	return r.inner.FindMany(ctx, name, exchangeName, symbol, unit, interval, indicatorDepth, depth, datetimes)
}

func (r *AnalysisRepository) FindManyByIndicator(
	ctx context.Context,
	exchangeName, symbol string,
	unit domain.Unit,
	interval int,
	datetime time.Time,
	indicatorDepth int,
	names []string,
	depths []int,
) ([]analysis.Analytic, error) {
	defer metrics.DBQueryHelper(r.db, "analysis_find_many_by_indicator")()
	return r.inner.FindManyByIndicator(ctx, exchangeName, symbol, unit, interval, datetime, indicatorDepth, names, depths)
}

func (r *AnalysisRepository) Last(
	ctx context.Context,
	exchangeName, symbol string,
	unit domain.Unit,
	interval int,
	name string,
	indicatorDepth, depth int,
) ([]analysis.Analytic, error) {
	defer metrics.DBQueryHelper(r.db, "analysis_last")()
	return r.inner.Last(ctx, exchangeName, symbol, unit, interval, name, indicatorDepth, depth)
}

func (r *AnalysisRepository) UniqGroups(ctx context.Context) ([]analysis.UniqGroup, error) {
	defer metrics.DBQueryHelper(r.db, "analysis_uniq_groups")()
	return r.inner.UniqGroups(ctx)
}

func (r *AnalysisRepository) LastInGroup(ctx context.Context, g analysis.UniqGroup, offset int) (*analysis.Analytic, error) {
	defer metrics.DBQueryHelper(r.db, "analysis_last_in_groups")()
	return r.inner.LastInGroup(ctx, g, offset)
}

func (r *AnalysisRepository) DeleteByGroup(ctx context.Context, g analysis.UniqGroup, datetime time.Time) error {
	defer metrics.DBQueryHelper(r.db, "analysis_delete_by_group")()
	return r.inner.DeleteByGroup(ctx, g, datetime)
}

func (r *AnalysisRepository) AllAnalyticInfo(ctx context.Context) (map[string][]analysis.AnalyticInfo, error) {
	defer metrics.DBQueryHelper(r.db, "analysis_all_analytic_info")()
	return r.inner.AllAnalyticInfo(ctx)
}
