package logging

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/util"
	"go.uber.org/zap"
)

var _ analysis.Repository = (*AnalysisRepository)(nil)

type AnalysisRepository struct {
	inner  analysis.Repository
	logger *zap.Logger
	db     string
}

func NewAnalysisRepository(inner analysis.Repository, logger *zap.Logger, db string) *AnalysisRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AnalysisRepository{inner: inner, logger: logger, db: db}
}

func (r *AnalysisRepository) Save(ctx context.Context, data ...analysis.Analytic) error {
	defer r.log(ctx, "analysis_save")()
	return r.inner.Save(ctx, data...)
}

func (r *AnalysisRepository) Find(
	ctx context.Context,
	name, exchangeName, symbol string,
	unit domain.Unit,
	datetime time.Time,
	interval, indicatorDepth, depth int,
) (*analysis.Analytic, error) {
	defer r.log(ctx, "analysis_find")()
	return r.inner.Find(ctx, name, exchangeName, symbol, unit, datetime, interval, indicatorDepth, depth)
}

func (r *AnalysisRepository) FindMany(
	ctx context.Context,
	name, exchangeName, symbol string,
	unit domain.Unit,
	interval, indicatorDepth, depth int,
	datetimes []time.Time,
) ([]analysis.Analytic, error) {
	defer r.log(ctx, "analysis_find_many")()
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
	defer r.log(ctx, "analysis_find_many_by_indicator")()
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
	defer r.log(ctx, "analysis_last")()
	return r.inner.Last(ctx, exchangeName, symbol, unit, interval, name, indicatorDepth, depth)
}

func (r *AnalysisRepository) UniqGroups(ctx context.Context) ([]analysis.UniqGroup, error) {
	defer r.log(ctx, "analysis_uniq_groups")()
	return r.inner.UniqGroups(ctx)
}

func (r *AnalysisRepository) LastInGroup(ctx context.Context, g analysis.UniqGroup, offset int) (*analysis.Analytic, error) {
	defer r.log(ctx, "analysis_last_in_groups")()
	return r.inner.LastInGroup(ctx, g, offset)
}

func (r *AnalysisRepository) DeleteByGroup(ctx context.Context, g analysis.UniqGroup, datetime time.Time) error {
	defer r.log(ctx, "analysis_delete_by_group")()
	return r.inner.DeleteByGroup(ctx, g, datetime)
}

func (r *AnalysisRepository) AllAnalyticInfo(ctx context.Context) (map[string][]analysis.AnalyticInfo, error) {
	defer r.log(ctx, "analysis_all_analytic_info")()
	return r.inner.AllAnalyticInfo(ctx)
}

func (r *AnalysisRepository) log(ctx context.Context, query string) func() {
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
