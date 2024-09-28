package calculators

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type macdHistogram struct {
	analyticService *analysis.Service
}

func NewMACDHistogram(analyticService *analysis.Service) analysis.CalculatorByAnalytic {
	return &macdHistogram{analyticService: analyticService}
}

func (m *macdHistogram) Name() string {
	return domain.MACDSHistogramIndicator
}

// Сигнальная линия уже формируется на основе основной линии, поэтому уже будет и основная и сигнальная
func (m *macdHistogram) ByAnalytic() string {
	return domain.MACDSignalLineIndicator
}

func (m *macdHistogram) SupportDepth(depth int) bool {
	return depth == 1
}

func (m *macdHistogram) SupportInterval(interval int) bool {
	return interval > 0
}

func (m *macdHistogram) Calculate(ctx context.Context, signalLine analysis.Analytic) ([]analysis.Analytic, error) {
	analyticData, err := m.analyticService.Analytics(
		ctx, signalLine.Exchange, signalLine.Symbol, signalLine.Unit, signalLine.Interval, domain.MACDMainLineIndicator, defaultLongDepthMACD, 1,
	)
	if err != nil {
		return nil, errors.Wrap(err, "fetch macd main line data")
	}
	var mainLine *analysis.Analytic
	for _, item := range analyticData {
		if item.Datetime.Equal(signalLine.Datetime) {
			mainLine = &item
			break
		}
	}
	if mainLine == nil {
		return nil, nil
	}
	return []analysis.Analytic{
		analysis.Analytic{
			ID:             uuid.New(),
			Name:           m.Name(),
			Exchange:       mainLine.Exchange,
			Symbol:         mainLine.Symbol,
			Unit:           mainLine.Unit,
			Interval:       mainLine.Interval,
			Datetime:       mainLine.Datetime,
			ByIndicator:    signalLine.Name,
			IndicatorDepth: mainLine.Depth,
			Depth:          1,
			Value:          mainLine.Value - signalLine.Value,
		},
	}, nil
}
