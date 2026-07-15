package analysis

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"time"
)

type UniqGroup struct {
	Name           string
	Exchange       string
	Symbol         string
	Unit           domain.Unit
	Interval       int
	Depth          int
	ByIndicator    string
	IndicatorDepth int
}

type AnalyticInfo struct {
	Unit           string
	Interval       int
	Name           string
	Depth          int
	IndicatorDepth int
}

type Repository interface {
	Save(ctx context.Context, data ...Analytic) error
	Last(
		ctx context.Context,
		exchangeName, symbol string,
		unit domain.Unit,
		interval int,
		name string,
		indicatorDepth, depth int,
	) ([]Analytic, error)
	Find(
		ctx context.Context,
		name, exchangeName, symbol string,
		unit domain.Unit,
		datetime time.Time,
		interval, indicatorDepth, depth int,
	) (*Analytic, error)
	// FindMany - Пакетный поиск по конкретным datetime, чтобы не ходить в БД по одной аналитике за раз.
	FindMany(
		ctx context.Context,
		name, exchangeName, symbol string,
		unit domain.Unit,
		interval, indicatorDepth, depth int,
		datetimes []time.Time,
	) ([]Analytic, error)
	// FindManyByIndicator - Пакетный поиск сразу по нескольким аналитикам (name) и глубинам (depth)
	// для одного индикатора (фиксированные exchange/symbol/unit/interval/datetime/indicatorDepth).
	FindManyByIndicator(
		ctx context.Context,
		exchangeName, symbol string,
		unit domain.Unit,
		interval int,
		datetime time.Time,
		indicatorDepth int,
		names []string,
		depths []int,
	) ([]Analytic, error)
	UniqGroups(ctx context.Context) ([]UniqGroup, error)
	LastInGroup(ctx context.Context, g UniqGroup, offset int) (*Analytic, error)
	DeleteByGroup(ctx context.Context, g UniqGroup, datetime time.Time) error
	AllAnalyticInfo(ctx context.Context) (map[string][]AnalyticInfo, error)
}
