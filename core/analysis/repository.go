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

type Repository interface {
	Save(ctx context.Context, data Analytic) error
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
	UniqGroups(ctx context.Context) ([]UniqGroup, error)
	LastInGroup(ctx context.Context, g UniqGroup, offset int) (*Analytic, error)
	DeleteByGroup(ctx context.Context, g UniqGroup, datetime time.Time) error
}
