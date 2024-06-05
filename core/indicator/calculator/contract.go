package calculator

import "github.com/AlekseyPorandaykin/crypto_polymath/domain"

type PrimaryIndicatorCalculator interface {
	Name() string
	SupportDepth(depth int) bool
	SupportInterval(interval int) bool
	Calculate(candlesticks []domain.Candlestick) *domain.Indicator
}

func calcExtremesCandlesticks(data []domain.Candlestick) (maxVal float64, minVal float64) {
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
