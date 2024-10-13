package calculators

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	defaultShortDepthMACD = 12
	defaultLongDepthMACD  = 26
)

type macdMainLine struct {
	indicatorService indicator.Indicator
}

func NewMACDMainLine(indicatorService indicator.Indicator) analysis.CalculatorByIndicator {
	return &macdMainLine{indicatorService: indicatorService}
}

func (m *macdMainLine) Name() string {
	return domain.MACDMainLineIndicator
}

func (m *macdMainLine) ByIndicator() string {
	return domain.EMAIndicator
}

// INFO: Если есть длинный EMA, то должен быть и короткий EMA
// Рассчитываем по двум последним ema: длинной и короткой
func (m *macdMainLine) SupportDepth(depth int) bool {
	return depth == 1
}

func (m *macdMainLine) SupportInterval(interval int) bool {
	return interval > 0
}

func (m *macdMainLine) Calculate(ctx context.Context, longEma domain.Indicator, depth int) (*analysis.Analytic, error) {
	if longEma.Depth != defaultLongDepthMACD {
		return nil, nil
	}
	shortEmaData, errShortEmaData := m.indicatorService.LastSequenceToDate(
		ctx,
		longEma.Exchange,
		longEma.Symbol,
		longEma.Unit,
		longEma.Interval,
		longEma.Name,
		defaultShortDepthMACD,
		1,
		longEma.Datetime)
	if errShortEmaData != nil {
		return nil, errors.Wrap(errShortEmaData, "fetch short ema data")
	}
	if len(shortEmaData) == 0 || shortEmaData[0].Datetime != longEma.Datetime {
		return nil, nil
	}
	return &analysis.Analytic{
		ID:             uuid.New(),
		Name:           m.Name(),
		Exchange:       longEma.Exchange,
		Symbol:         longEma.Symbol,
		Unit:           longEma.Unit,
		Interval:       longEma.Interval,
		Datetime:       longEma.Datetime,
		ByIndicator:    longEma.Name,
		IndicatorDepth: longEma.Depth,
		Depth:          1,
		Value:          shortEmaData[0].Value - longEma.Value,
	}, nil
}
