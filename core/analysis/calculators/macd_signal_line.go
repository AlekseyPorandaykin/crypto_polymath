package calculators

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/util"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const defaultDepthSmoothingMACD = 9

type macdSignalLine struct {
	analyticService *analysis.Service
}

func NewMACDSignalLine(analyticService *analysis.Service) analysis.CalculatorByAnalytic {
	return &macdSignalLine{analyticService: analyticService}
}

func (m *macdSignalLine) Name() string {
	return domain.MACDSignalLineIndicator
}

func (m *macdSignalLine) ByAnalytic() string {
	return domain.MACDMainLineIndicator
}

func (m *macdSignalLine) SupportDepth(depth int) bool {
	return depth > 0
}

func (m *macdSignalLine) SupportInterval(interval int) bool {
	return interval > 0
}

func (m *macdSignalLine) Calculate(ctx context.Context, mainLine analysis.Analytic) ([]analysis.Analytic, error) {
	analyticData, err := m.analyticService.SequenceAnalytics(
		ctx, mainLine.Exchange, mainLine.Symbol, mainLine.Unit, mainLine.Interval, mainLine.Name, mainLine.IndicatorDepth, mainLine.Depth,
	)
	if err != nil {
		return nil, errors.Wrap(err, "fetch macd main line data")
	}
	emaData := make([]analysis.Analytic, 0, defaultDepthSmoothingMACD)
	for i := range analyticData {
		if i >= defaultDepthSmoothingMACD {
			break
		}
		emaData = append(emaData, analyticData[i])
	}
	if len(emaData) != defaultDepthSmoothingMACD {
		return nil, nil
	}
	slice.SortBy[analysis.Analytic](emaData, func(a, b analysis.Analytic) bool {
		return a.Datetime.Before(b.Datetime)
	})
	emaVal := domain.EMA(util.ModifySlice(emaData, func(item analysis.Analytic) float64 {
		return item.Value
	}))
	return []analysis.Analytic{
		analysis.Analytic{
			ID:             uuid.New(),
			Name:           m.Name(),
			Exchange:       mainLine.Exchange,
			Symbol:         mainLine.Symbol,
			Unit:           mainLine.Unit,
			Interval:       mainLine.Interval,
			Datetime:       mainLine.Datetime,
			ByIndicator:    mainLine.Name,
			IndicatorDepth: mainLine.Depth,
			Depth:          defaultDepthSmoothingMACD,
			Value:          emaVal,
		},
	}, nil
}
