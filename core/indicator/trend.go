package indicator

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/util"
	"github.com/duke-git/lancet/v2/slice"
)

const trendThreshold = 60

type trend struct {
}

func NewTrend() Calculator {
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
	for _, batchCandles := range util.BatchSlice(data, lenBatch(len(data))) {
		maxVal, minVal := calcExtremes(batchCandles)
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

func calcTrend(maxValues, minValues []float64) int {
	var countUpward, countDownward int
	var prevMax, prevMin, percentUpward, percentDownward float64
	for _, maxValue := range maxValues {
		if prevMax != 0 && maxValue > prevMax {
			countUpward++
		}
		prevMax = maxValue
	}
	for _, minValue := range minValues {
		if prevMin != 0 && prevMin > minValue {
			countDownward++
		}
		prevMin = minValue
	}
	if countUpward > 0 {
		percentUpward = (float64(countUpward) / float64(len(maxValues))) * 100
	}
	if countDownward > 0 {
		percentDownward = (float64(countDownward) / float64(len(minValues))) * 100
	}
	if countUpward == countDownward {
		return domain.FlatTrend
	}
	if countDownward == 0 && countUpward != 0 && percentUpward >= trendThreshold {
		return domain.UpwardTrend
	}
	if countUpward == 0 && countDownward != 0 && percentDownward >= trendThreshold {
		return domain.DownwardTrend
	}

	if percentUpward > percentDownward && percentUpward >= trendThreshold {
		return domain.UpwardTrend
	}
	if percentDownward > percentUpward && percentDownward >= trendThreshold {
		return domain.DownwardTrend
	}

	return domain.FlatTrend
}

func calcExtremes(data []domain.Candlestick) (maxVal float64, minVal float64) {
	for _, item := range data {
		if item.ClosePrice > maxVal {
			maxVal = item.ClosePrice
		}
		if minVal == 0 || minVal > item.ClosePrice {
			minVal = item.ClosePrice
		}
	}

	return maxVal, minVal
}

func lenBatch(count int) int {
	if count <= 15 {
		return 3
	}
	if count <= 20 {
		return 4
	}
	if count < 50 {
		return 5
	}
	return 10
}
