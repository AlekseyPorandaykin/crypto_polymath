package analysis

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/google/uuid"
	"time"
)

type Analytic struct {
	ID             uuid.UUID
	Exchange       string
	Symbol         string
	Unit           domain.Unit
	Interval       int
	Name           string
	Datetime       time.Time
	Depth          int
	ByIndicator    string
	IndicatorDepth int
	Value          float64
}

type CalculatorByIndicator interface {
	Name() string
	ByIndicator() string
	SupportDepth(depth int) bool
	SupportInterval(interval int) bool
	Calculate(ctx context.Context, indicator domain.Indicator) ([]Analytic, error)
}

type CalculatorByAnalytic interface {
	Name() string
	ByAnalytic() string
	SupportDepth(depth int) bool
	SupportInterval(interval int) bool
	Calculate(ctx context.Context, data Analytic) ([]Analytic, error)
}
