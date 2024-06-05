package calculator

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/duke-git/lancet/v2/slice"
)

type TypeCandle struct {
}

func NewTypeCandle() PrimaryIndicatorCalculator {
	return &TypeCandle{}
}

func (t *TypeCandle) Name() string {
	return domain.TypeCandleIndicator
}

func (t *TypeCandle) SupportDepth(depth int) bool {
	return depth == 1
}

func (t *TypeCandle) SupportInterval(interval int) bool {
	return interval > 0
}

func (t *TypeCandle) Calculate(data []domain.Candlestick) *domain.Indicator {
	if len(data) < 1 {
		return nil
	}
	slice.SortBy[domain.Candlestick](data, func(a, b domain.Candlestick) bool {
		return a.StartTime.After(b.StartTime)
	})
	candlestick := data[0]
	val := domain.UpCandle
	if candlestick.ClosePrice < candlestick.OpenPrice {
		val = domain.DownCandle
	}

	indicator := domain.Indicator{
		Symbol:   candlestick.Symbol,
		Exchange: candlestick.Exchange,
		Unit:     candlestick.Unit,
		Interval: candlestick.Interval,
		Name:     domain.TypeCandleIndicator,
		Depth:    len(data),
		Datetime: candlestick.StartTime,
		Value:    float64(val),
	}

	return &indicator
}
