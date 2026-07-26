package calculators

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/go-kit/pkg/util"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/google/uuid"
)

const trendThreshold = 60

type trendByEMA struct {
	indicatorService indicator.Indicator
}

func NewTrendByEMA(indicatorService indicator.Indicator) analysis.CalculatorByIndicator {
	return &trendByEMA{indicatorService: indicatorService}
}

func (t *trendByEMA) Name() string {
	return domain.TrendByEMAIndicator
}

func (t *trendByEMA) ByIndicator() string {
	return domain.EMAIndicator
}

func (t *trendByEMA) SupportDepth(depth int) bool {
	return depth >= 10
}

func (t *trendByEMA) SupportInterval(interval int) bool {
	return interval > 0
}

func (t *trendByEMA) Calculate(ctx context.Context, indicatorData domain.Indicator, depth int) (*analysis.Analytic, error) {
	data, err := t.indicatorService.LastSequenceToDate(
		ctx,
		indicatorData.Exchange,
		indicatorData.Symbol,
		indicatorData.Unit,
		indicatorData.Interval,
		indicatorData.Name,
		indicatorData.Depth,
		depth,
		indicatorData.Datetime,
	)
	if err != nil {
		return nil, err
	}
	emaData := util.ClearSlice[domain.Indicator](data, func(item domain.Indicator) bool {
		return item.Name == t.ByIndicator()
	})
	if len(emaData) == 0 && len(emaData) != depth {
		return nil, nil
	}
	//
	slice.SortBy[domain.Indicator](emaData, func(a, b domain.Indicator) bool {
		return a.Datetime.Before(b.Datetime)
	})
	maxValues := make([]float64, 0, len(emaData))
	minValues := make([]float64, 0, len(emaData))
	batches := util.BatchSlice[domain.Indicator](emaData, lenBatch(len(emaData)))
	for _, batch := range batches {
		maxVal, minVal := calcExtremesIndicators(batch)
		maxValues = append(maxValues, maxVal)
		minValues = append(minValues, minVal)
	}

	return &analysis.Analytic{
		ID:             uuid.New(),
		Name:           t.Name(),
		Exchange:       indicatorData.Exchange,
		Symbol:         indicatorData.Symbol,
		Unit:           indicatorData.Unit,
		Interval:       indicatorData.Interval,
		Datetime:       indicatorData.Datetime,
		ByIndicator:    indicatorData.Name,
		IndicatorDepth: indicatorData.Depth,
		Depth:          depth,
		Value:          float64(calcTrend(maxValues, minValues)),
	}, nil
}

func calcExtremesIndicators(data []domain.Indicator) (maxVal float64, minVal float64) {
	for _, item := range data {
		if item.Value > maxVal {
			maxVal = item.Value
		}
		if minVal == 0 || minVal > item.Value {
			minVal = item.Value
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
