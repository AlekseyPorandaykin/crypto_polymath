package repository

import (
	"context"
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
