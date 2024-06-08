package memory

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/metrics"
	"github.com/duke-git/lancet/v2/slice"
	lru "github.com/hashicorp/golang-lru/v2"
	"sync"
	"time"
)

type AnalysisRepository struct {
	caches map[string]*lru.Cache[time.Time, analysis.Analytic]
	mu     sync.Mutex
	size   int
}

func NewAnalysisRepository(size int) analysis.Repository {
	return &AnalysisRepository{
		caches: make(map[string]*lru.Cache[time.Time, analysis.Analytic]),
		size:   size,
	}
}

func (a *AnalysisRepository) Save(ctx context.Context, data analysis.Analytic) error {
	defer metrics.CacheQueryHelper("memory", "analysis_save")()
	a.mu.Lock()
	defer a.mu.Unlock()
	key := createKey(data.Exchange, data.Symbol, data.Unit, data.Interval, data.Name, data.IndicatorDepth, data.Depth)
	if a.caches[key] == nil {
		cache, err := lru.New[time.Time, analysis.Analytic](a.size)
		if err != nil {
			return err
		}
		a.caches[key] = cache
	}
	a.caches[key].Add(data.Datetime, data)

	return nil
}

func (a *AnalysisRepository) Last(
	ctx context.Context, exchangeName, symbol string, unit domain.Unit, interval int, name string, indicatorDepth, depth int,
) ([]analysis.Analytic, error) {
	defer metrics.CacheQueryHelper("memory", "analysis_last")()
	a.mu.Lock()
	defer a.mu.Unlock()
	key := createKey(exchangeName, symbol, unit, interval, name, indicatorDepth, depth)
	cache, has := a.caches[key]
	if !has {
		return nil, nil
	}
	data := cache.Values()
	slice.SortBy[analysis.Analytic](data, func(a, b analysis.Analytic) bool {
		return a.Datetime.After(b.Datetime)
	})
	return data, nil
}

func (a *AnalysisRepository) UniqGroups(ctx context.Context) ([]analysis.UniqGroup, error) {
	defer metrics.CacheQueryHelper("memory", "analysis_uniq_groups")()
	a.mu.Lock()
	defer a.mu.Unlock()
	data := make([]analysis.UniqGroup, 0, 1_000)
	uniqData := make(map[analysis.UniqGroup]struct{})
	for _, cache := range a.caches {
		for _, item := range cache.Values() {
			uq := analysis.UniqGroup{
				Name:           item.Name,
				Exchange:       item.Exchange,
				Symbol:         item.Symbol,
				Unit:           item.Unit,
				Interval:       item.Interval,
				Depth:          item.Depth,
				ByIndicator:    item.ByIndicator,
				IndicatorDepth: item.IndicatorDepth,
			}
			if _, has := uniqData[uq]; has {
				continue
			}
			data = append(data, uq)
			uniqData[uq] = struct{}{}
		}
	}
	return data, nil
}

func (a *AnalysisRepository) LastInGroup(ctx context.Context, g analysis.UniqGroup, offset int) (*analysis.Analytic, error) {
	return nil, nil
}
func (a *AnalysisRepository) DeleteByGroup(ctx context.Context, g analysis.UniqGroup, datetime time.Time) error {
	return nil
}

func createKey(exchangeName, symbol string, unit domain.Unit, interval int, name string, indicatorDepth, depth int) string {
	return fmt.Sprintf("%s-%s-%s-%d-%s-%d-%d", exchangeName, symbol, unit, interval, name, indicatorDepth, depth)
}
