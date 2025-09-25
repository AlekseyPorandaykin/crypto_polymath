package repository

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

type AnalysisRepository struct {
	storage analysis.Repository
	cache   analysis.Repository
}

func NewAnalysisRepository(storage, cache analysis.Repository) analysis.Repository {
	return &AnalysisRepository{storage: storage, cache: cache}
}

func (a *AnalysisRepository) Save(ctx context.Context, data analysis.Analytic) error {
	_ = a.cache.Save(ctx, data)
	return a.storage.Save(ctx, data)
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
	for _, itemStorage := range dataStorage {
		_ = a.cache.Save(ctx, itemStorage)
	}
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
