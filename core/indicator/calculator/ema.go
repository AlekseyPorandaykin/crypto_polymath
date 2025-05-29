package calculator

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/go-template/pkg/util"
	"github.com/duke-git/lancet/v2/slice"
)

type ema struct {
}

func NewEMA() PrimaryIndicatorCalculator {
	return &ema{}
}

func (e *ema) Name() string {
	return domain.EMAIndicator
}

func (e *ema) SupportDepth(depth int) bool {
	return depth > 1
}

func (e *ema) SupportInterval(interval int) bool {
	return interval > 0
}
func (e *ema) Calculate(data []domain.Candlestick) *domain.Indicator {
	slice.SortBy[domain.Candlestick](data, func(a, b domain.Candlestick) bool {
		return a.StartTime.Before(b.StartTime)
	})
	candlestick := data[len(data)-1]
	emaVal := domain.EMA(util.ModifySlice(data, func(item domain.Candlestick) float64 {
		return item.ClosePrice
	}))
	indicator := domain.Indicator{
		Symbol:   candlestick.Symbol,
		Exchange: candlestick.Exchange,
		Unit:     candlestick.Unit,
		Interval: candlestick.Interval,
		Name:     domain.EMAIndicator,
		Depth:    len(data),
		Datetime: candlestick.StartTime,
		Value:    emaVal,
	}
	return &indicator
}
