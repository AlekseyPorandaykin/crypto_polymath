package calculators

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/google/uuid"
)

const defaultDepthStochasticSignalLine = 3

type stochasticSignalLine struct {
	indicatorService indicator.Indicator
}

func NewStochasticSignalLine(indicatorService indicator.Indicator) analysis.CalculatorByIndicator {
	return &stochasticSignalLine{indicatorService: indicatorService}
}

func (s *stochasticSignalLine) Name() string {
	return domain.StochasticSignalLine
}

func (s *stochasticSignalLine) ByIndicator() string {
	return domain.StochasticMainLine
}

func (s *stochasticSignalLine) SupportDepth(depth int) bool {
	return depth == defaultDepthStochasticSignalLine
}

func (s *stochasticSignalLine) SupportInterval(interval int) bool {
	return interval > 0
}

func (s *stochasticSignalLine) Calculate(ctx context.Context, indicatorData domain.Indicator) ([]analysis.Analytic, error) {
	data, err := s.indicatorService.LastToDate(
		ctx,
		indicatorData.Exchange,
		indicatorData.Symbol,
		indicatorData.Unit,
		indicatorData.Interval,
		indicatorData.Name,
		indicatorData.Depth,
		defaultDepthStochasticSignalLine,
		indicatorData.Datetime,
	)
	if err != nil {
		return nil, err
	}
	var sumMainLineVal float64
	for _, item := range data {
		sumMainLineVal += item.Value
	}

	return []analysis.Analytic{
		analysis.Analytic{
			ID:             uuid.New(),
			Name:           s.Name(),
			Exchange:       indicatorData.Exchange,
			Symbol:         indicatorData.Symbol,
			Unit:           indicatorData.Unit,
			Interval:       indicatorData.Interval,
			Datetime:       indicatorData.Datetime,
			ByIndicator:    indicatorData.Name,
			IndicatorDepth: indicatorData.Depth,
			Depth:          defaultDepthStochasticSignalLine,
			Value:          sumMainLineVal / float64(len(data)),
		},
	}, nil
}
