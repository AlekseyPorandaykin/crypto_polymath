package calculator

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/duke-git/lancet/v2/slice"
)

const defaultDepthStochasticMainLine = 14

type stochasticMainLine struct {
}

func NewStochasticMainLine() PrimaryIndicatorCalculator {
	return &stochasticMainLine{}
}

func (s stochasticMainLine) Name() string {
	return domain.StochasticMainLine
}

func (s stochasticMainLine) SupportDepth(depth int) bool {
	return depth > 1
}

func (s stochasticMainLine) SupportInterval(interval int) bool {
	return interval > 0
}

func (s stochasticMainLine) Calculate(data []domain.Candlestick) *domain.Indicator {
	if !s.SupportDepth(len(data)) {
		return nil
	}
	slice.SortBy[domain.Candlestick](data, func(a, b domain.Candlestick) bool {
		return a.StartTime.Before(b.StartTime)
	})
	candlestick := data[len(data)-1]
	var maxVal, minVal float64
	for _, item := range data {
		if item.ClosePrice > maxVal {
			maxVal = item.ClosePrice
		}
		if item.ClosePrice < minVal || minVal == 0 {
			minVal = item.ClosePrice
		}
	}
	// (Цена в выбранной точке − Минимальная цена) / (Максимальная цена − Минимальная цена) × 100%
	var value float64
	if maxVal != minVal {
		value = (candlestick.ClosePrice - minVal) / (maxVal - minVal) * 100
	}
	indicator := domain.Indicator{
		Symbol:   candlestick.Symbol,
		Exchange: candlestick.Exchange,
		Unit:     candlestick.Unit,
		Interval: candlestick.Interval,
		Name:     s.Name(),
		Depth:    len(data),
		Datetime: candlestick.StartTime,
		Value:    value,
	}
	return &indicator
}
