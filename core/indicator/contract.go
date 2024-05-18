package indicator

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"time"
)

type Candlestick interface {
	LastCandlesticks(
		ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int, datetime time.Time,
	) ([]domain.Candlestick, error)
	NextCandlesticks(
		ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int, datetime time.Time,
	) ([]domain.Candlestick, error)

	FirstCandlestick(
		ctx context.Context, exchange, symbol string, unit domain.Unit, interval int, offset int,
	) (*domain.Candlestick, error)
}

type Calculator interface {
	Name() string
	SupportDepth(depth int) bool
	SupportInterval(interval int) bool
	Calculate(candlesticks []domain.Candlestick) *domain.Indicator
}

type Indicator interface {
	AddCalculator(calculator Calculator)
	Indicator(ctx context.Context, candlestick domain.Candlestick, name string, depth int) (*domain.Indicator, error)
	Indicators(ctx context.Context, exchange, symbol string, unit domain.Unit, interval int, name string, depth, limit int) ([]domain.Indicator, error)
	CalcIndicators(ctx context.Context, exchange, symbol string, unit domain.Unit, interval, depth int) (int, error)
	DeleteOldRows(ctx context.Context, oldValueLimit int) error
}
