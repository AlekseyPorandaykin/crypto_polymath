package calculator

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/util"
	"github.com/duke-git/lancet/v2/slice"
)

const trendThreshold = 60

type trend struct {
}

func NewTrend() PrimaryIndicatorCalculator {
	return &trend{}
}

func (t trend) Name() string {
	return domain.TrendIndicator
}

func (t trend) SupportDepth(depth int) bool {
	return depth >= 10
}

func (t trend) SupportInterval(interval int) bool {
	return interval > 0
}

func (t trend) Calculate(data []domain.Candlestick) *domain.Indicator {
	maxValues := make([]float64, 0, len(data))
	minValues := make([]float64, 0, len(data))
	candlestick := data[0]
	slice.SortBy[domain.Candlestick](data, func(a, b domain.Candlestick) bool {
		return a.StartTime.Before(b.StartTime)
	})
	batches := util.BatchSlice[domain.Candlestick](data, lenBatch(len(data)))
	for _, batchCandles := range batches {
		maxVal, minVal := calcExtremesCandlesticks(batchCandles)
		maxValues = append(maxValues, maxVal)
		minValues = append(minValues, minVal)
	}

	indicator := domain.Indicator{
		Symbol:   candlestick.Symbol,
		Exchange: candlestick.Exchange,
		Unit:     candlestick.Unit,
		Interval: candlestick.Interval,
		Name:     domain.TrendIndicator,
		Depth:    len(data),
		Datetime: candlestick.StartTime,
		Value:    float64(calcTrend(maxValues, minValues)),
	}

	return &indicator
}
