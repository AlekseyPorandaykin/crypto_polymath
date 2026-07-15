package repository

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

var _ analysis.Repository = (*AnalysisRepository)(nil)

type AnalysisRepository struct {
	storage analysis.Repository
	cache   analysis.Repository
}

func NewAnalysisRepository(storage, cache analysis.Repository) analysis.Repository {
	return &AnalysisRepository{storage: storage, cache: cache}
}

func (a *AnalysisRepository) Save(ctx context.Context, data ...analysis.Analytic) error {
	_ = a.cache.Save(ctx, data...)
	return a.storage.Save(ctx, data...)
}

func (a *AnalysisRepository) Find(
	ctx context.Context,
	name, exchangeName, symbol string,
	unit domain.Unit,
	datetime time.Time,
	interval, indicatorDepth, depth int,
) (*analysis.Analytic, error) {
	dataCache, err := a.cache.Find(ctx, name, exchangeName, symbol, unit, datetime, interval, indicatorDepth, depth)
	if err != nil {
		return nil, err
	}
	if dataCache != nil {
		return dataCache, nil
	}
	dataStorage, err := a.storage.Find(ctx, name, exchangeName, symbol, unit, datetime, interval, indicatorDepth, depth)
	if err != nil {
		return nil, err
	}
	if dataStorage != nil {
		_ = a.cache.Save(ctx, *dataStorage)
	}
	return dataStorage, nil
}

func (a *AnalysisRepository) FindMany(
	ctx context.Context,
	name, exchangeName, symbol string,
	unit domain.Unit,
	interval, indicatorDepth, depth int,
	datetimes []time.Time,
) ([]analysis.Analytic, error) {
	if len(datetimes) == 0 {
		return nil, nil
	}
	cacheData, err := a.cache.FindMany(ctx, name, exchangeName, symbol, unit, interval, indicatorDepth, depth, datetimes)
	if err != nil {
		return nil, err
	}
	found := make(map[time.Time]struct{}, len(cacheData))
	for _, item := range cacheData {
		found[item.Datetime] = struct{}{}
	}
	missing := make([]time.Time, 0, len(datetimes)-len(cacheData))
	for _, datetime := range datetimes {
		if _, ok := found[datetime]; !ok {
			missing = append(missing, datetime)
		}
	}
	if len(missing) == 0 {
		return cacheData, nil
	}
	dataStorage, err := a.storage.FindMany(ctx, name, exchangeName, symbol, unit, interval, indicatorDepth, depth, missing)
	if err != nil {
		return nil, err
	}
	if len(dataStorage) > 0 {
		_ = a.cache.Save(ctx, dataStorage...)
	}
	return append(cacheData, dataStorage...), nil
}
func (a *AnalysisRepository) FindManyByIndicator(
	ctx context.Context,
	exchangeName, symbol string,
	unit domain.Unit,
	interval int,
	datetime time.Time,
	indicatorDepth int,
	names []string,
	depths []int,
) ([]analysis.Analytic, error) {
	if len(names) == 0 || len(depths) == 0 {
		return nil, nil
	}
	cacheData, err := a.cache.FindManyByIndicator(ctx, exchangeName, symbol, unit, interval, datetime, indicatorDepth, names, depths)
	if err != nil {
		return nil, err
	}
	// В кеше не различить "ещё не посчитано" от "не закешировано" - если нашли не всё,
	// перезапрашиваем весь набор из storage (он источник истины) и прогреваем кеш заново.
	if len(cacheData) >= len(names)*len(depths) {
		return cacheData, nil
	}
	dataStorage, err := a.storage.FindManyByIndicator(ctx, exchangeName, symbol, unit, interval, datetime, indicatorDepth, names, depths)
	if err != nil {
		return nil, err
	}
	if len(dataStorage) > 0 {
		_ = a.cache.Save(ctx, dataStorage...)
	}
	return dataStorage, nil
}

func (a *AnalysisRepository) Last(ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int, name string, indicatorDepth, depth int) ([]analysis.Analytic, error) {
	dataCache, err := a.cache.Last(ctx, exchangeName, symbol, unit, interval, name, indicatorDepth, depth)
	if err != nil {
		return nil, err
	}
	if len(dataCache) > 0 {
		return dataCache, nil
	}
	dataStorage, err := a.storage.Last(ctx, exchangeName, symbol, unit, interval, name, indicatorDepth, depth)
	if err != nil {
		return nil, err
	}
	if len(dataStorage) == 0 {
		return nil, nil
	}
	_ = a.cache.Save(ctx, dataStorage...)
	return dataStorage, nil
}

func (a *AnalysisRepository) UniqGroups(ctx context.Context) ([]analysis.UniqGroup, error) {
	return a.storage.UniqGroups(ctx)
}
func (a *AnalysisRepository) DeleteByGroup(ctx context.Context, g analysis.UniqGroup, datetime time.Time) error {
	return a.storage.DeleteByGroup(ctx, g, datetime)
}

func (a *AnalysisRepository) LastInGroup(ctx context.Context, g analysis.UniqGroup, offset int) (*analysis.Analytic, error) {
	return a.storage.LastInGroup(ctx, g, offset)
}

func (a *AnalysisRepository) AllAnalyticInfo(ctx context.Context) (map[string][]analysis.AnalyticInfo, error) {
	return a.storage.AllAnalyticInfo(ctx)
}
