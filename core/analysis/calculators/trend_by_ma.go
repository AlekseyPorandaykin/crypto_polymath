package calculators

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/util"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/google/uuid"
)

type trendByMA struct {
	indicatorService indicator.Indicator
	depths           []int
}

func NewTrendByMA(indicatorService indicator.Indicator, depths []int) analysis.CalculatorByIndicator {
	return &trendByMA{indicatorService: indicatorService, depths: depths}
}

func (t *trendByMA) Name() string {
	return domain.TrendByMAIndicator
}

func (t *trendByMA) ByIndicator() string {
	return domain.MAIndicator
}

func (t *trendByMA) SupportDepth(depth int) bool {
	return depth >= 10
}

func (t *trendByMA) SupportInterval(interval int) bool {
	return interval > 0
}

func (t *trendByMA) Calculate(ctx context.Context, indicatorData domain.Indicator) ([]analysis.Analytic, error) {
	analytics := make([]analysis.Analytic, 0, len(t.depths))
	for _, depth := range t.depths {
		if !t.SupportDepth(depth) {
			continue
		}
		analyticData, err := t.calculateForDepth(ctx, indicatorData, depth)
		if err != nil {
			return nil, err
		}
		if analyticData == nil {
			continue
		}
		analytics = append(analytics, *analyticData)
	}
	return analytics, nil
}

func (t *trendByMA) calculateForDepth(ctx context.Context, indicatorData domain.Indicator, depth int) (*analysis.Analytic, error) {
	data, err := t.indicatorService.LastToDate(
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
	maData := util.ClearSlice[domain.Indicator](data, func(item domain.Indicator) bool {
		return item.Name == t.ByIndicator()
	})
	if len(maData) == 0 && len(maData) != depth {
		return nil, nil
	}

	slice.SortBy[domain.Indicator](maData, func(a, b domain.Indicator) bool {
		return a.Datetime.Before(b.Datetime)
	})
	maxValues := make([]float64, 0, len(maData))
	minValues := make([]float64, 0, len(maData))
	batches := util.BatchSlice[domain.Indicator](maData, lenBatch(len(maData)))
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
