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

type rsi struct {
	indicatorService indicator.Indicator
}

func NewRSI(indicatorService indicator.Indicator) analysis.CalculatorByIndicator {
	return &rsi{indicatorService}
}

func (r *rsi) Name() string {
	return domain.RSIIndicator
}

func (r *rsi) ByIndicator() string {
	return domain.EMAIndicator
}

func (r *rsi) SupportDepth(depth int) bool {
	return depth >= 10
}

func (r *rsi) SupportInterval(interval int) bool {
	return interval > 0
}

func (r *rsi) Calculate(ctx context.Context, indicatorData domain.Indicator, depth int) (*analysis.Analytic, error) {
	data, err := r.indicatorService.LastSequenceToDate(
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
		return item.Name == r.ByIndicator()
	})
	if len(emaData) == 0 && len(emaData) != depth {
		return nil, nil
	}
	slice.SortBy[domain.Indicator](emaData, func(a, b domain.Indicator) bool {
		return a.Datetime.Before(b.Datetime)
	})
	var (
		growthEMA float64 //рост
		fallEMA   float64 //падение
		prevVal   float64
	)
	for _, item := range emaData {
		if prevVal == 0 {
			prevVal = item.Value
			continue
		}
		if item.Value > prevVal {
			growthEMA += item.Value - prevVal
			continue
		}
		fallEMA += prevVal - item.Value
		prevVal = item.Value
	}
	var relativeStrength float64
	if growthEMA != 0 && fallEMA != 0 {
		relativeStrength = growthEMA / fallEMA
	}
	val := 100 - (100 / (1 + relativeStrength))
	return &analysis.Analytic{
		ID:             uuid.New(),
		Name:           r.Name(),
		Exchange:       indicatorData.Exchange,
		Symbol:         indicatorData.Symbol,
		Unit:           indicatorData.Unit,
		Interval:       indicatorData.Interval,
		Datetime:       indicatorData.Datetime,
		ByIndicator:    indicatorData.Name,
		IndicatorDepth: indicatorData.Depth,
		Depth:          depth,
		Value:          val,
	}, nil
}
