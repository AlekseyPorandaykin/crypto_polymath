package calculator

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/duke-git/lancet/v2/slice"
	"math"
)

type PriceChanges struct {
}

func NewPriceChanges() PrimaryIndicatorCalculator {
	return &PriceChanges{}
}

func (p *PriceChanges) Name() string {
	return domain.PriceChanges
}

func (p *PriceChanges) SupportDepth(depth int) bool {
	return depth > 1
}

func (p *PriceChanges) SupportInterval(interval int) bool {
	return interval > 0
}

func (p *PriceChanges) Calculate(data []domain.Candlestick) *domain.Indicator {
	if len(data) < 1 {
		return nil
	}
	slice.SortBy[domain.Candlestick](data, func(a, b domain.Candlestick) bool {
		return a.StartTime.After(b.StartTime)
	})
	candlestick := data[0]
	var prevValue, changes float64
	for _, item := range data {
		if prevValue == 0 {
			prevValue = item.ClosePrice
			continue
		}
		changes += math.Abs(prevValue - item.ClosePrice)
		prevValue = item.ClosePrice
	}
	indicator := domain.Indicator{
		Symbol:   candlestick.Symbol,
		Exchange: candlestick.Exchange,
		Unit:     candlestick.Unit,
		Interval: candlestick.Interval,
		Name:     domain.PriceChanges,
		Depth:    len(data),
		Datetime: candlestick.StartTime,
		Value:    float64(int((changes/float64(len(data))/candlestick.ClosePrice)*1_000_000)) / 1_000_000,
	}

	return &indicator
}
