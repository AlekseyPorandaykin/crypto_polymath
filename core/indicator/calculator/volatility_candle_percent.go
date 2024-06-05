package calculator

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/duke-git/lancet/v2/slice"
)

type VolatilityCandlePercent struct {
}

func NewVolatilityCandlePercent() PrimaryIndicatorCalculator {
	return &VolatilityCandlePercent{}
}
func (v *VolatilityCandlePercent) SupportDepth(depth int) bool {
	return depth == 1
}

func (v *VolatilityCandlePercent) SupportInterval(interval int) bool {
	return interval > 0
}

func (v *VolatilityCandlePercent) Name() string {
	return domain.VolatilityCandlePercentIndicator
}

func (v *VolatilityCandlePercent) Calculate(data []domain.Candlestick) *domain.Indicator {
	if len(data) < 1 {
		return nil
	}
	slice.SortBy[domain.Candlestick](data, func(a, b domain.Candlestick) bool {
		return a.StartTime.After(b.StartTime)
	})
	candlestick := data[0]
	val := (candlestick.OpenPrice - candlestick.ClosePrice) / candlestick.ClosePrice * 100
	if candlestick.ClosePrice > candlestick.OpenPrice {
		val = (candlestick.ClosePrice - candlestick.OpenPrice) / candlestick.OpenPrice * 100
	}
	indicator := domain.Indicator{
		Symbol:   candlestick.Symbol,
		Exchange: candlestick.Exchange,
		Unit:     candlestick.Unit,
		Interval: candlestick.Interval,
		Name:     v.Name(),
		Depth:    len(data),
		Datetime: candlestick.StartTime,
		Value:    float64(int(val*1_000_000)) / 1_000_000,
	}

	return &indicator
}
