package indicator

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/duke-git/lancet/v2/slice"
)

type ma struct {
}

func NewMA() Calculator {
	return &ma{}
}

func (m *ma) Name() string {
	return domain.MAIndicator
}

func (m *ma) SupportDepth(depth int) bool {
	return depth > 1
}

func (m *ma) SupportInterval(interval int) bool {
	return interval > 0
}
func (m *ma) Calculate(data []domain.Candlestick) *domain.Indicator {
	if len(data) < 1 {
		return nil
	}
	slice.SortBy[domain.Candlestick](data, func(a, b domain.Candlestick) bool {
		return a.StartTime.After(b.StartTime)
	})
	candlestick := data[0]
	var sumValues float64
	for _, item := range data {
		sumValues += item.ClosePrice
	}
	indicator := domain.Indicator{
		Symbol:   candlestick.Symbol,
		Exchange: candlestick.Exchange,
		Unit:     candlestick.Unit,
		Interval: candlestick.Interval,
		Name:     domain.MAIndicator,
		Depth:    len(data),
		Datetime: candlestick.StartTime,
		Value:    float64(int(sumValues/float64(len(data))*1_000)) / 1_000,
	}

	return &indicator
}
