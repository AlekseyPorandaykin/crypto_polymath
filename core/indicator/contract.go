package indicator

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator/calculator"
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

type Indicator interface {
	AddCalculator(calculator calculator.PrimaryIndicatorCalculator)
	//Indicators - получить индикаторы, которые есть в бд
	Indicators(ctx context.Context, exchange, symbol string, unit domain.Unit, interval int, name string, depth, limit int) ([]domain.Indicator, error)
	LastSequenceToDate(ctx context.Context, exchange, symbol string, unit domain.Unit, interval int, name string, depth, limit int, to time.Time) ([]domain.Indicator, error)
	CalcIndicators(ctx context.Context, exchange, symbol string, unit domain.Unit, interval, depth int) ([]domain.Indicator, error)
	CalcIndicatorsByCandlestick(ctx context.Context, candlestick domain.Candlestick, depth int) ([]domain.Indicator, error)
	DeleteOldRows(ctx context.Context, oldValueLimit int) error
}
