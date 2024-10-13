package calculators

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/google/uuid"
)

type ratioCandleToMA struct {
	candleService candlestick.Candlestick
}

func NewRationCandleToMA(candleService candlestick.Candlestick) analysis.CalculatorByIndicator {
	return &ratioCandleToMA{candleService: candleService}
}

func (c *ratioCandleToMA) Name() string {
	return domain.RatioCandleToMAIndicator
}

func (c *ratioCandleToMA) ByIndicator() string {
	return domain.MAIndicator
}

func (c *ratioCandleToMA) SupportDepth(depth int) bool {
	return depth == 1
}

func (c *ratioCandleToMA) SupportInterval(interval int) bool {
	return interval > 0
}

func (c *ratioCandleToMA) Calculate(ctx context.Context, indicatorData domain.Indicator, depth int) (*analysis.Analytic, error) {
	candle, err := c.candleByIndicator(ctx, indicatorData)
	if err != nil {
		return nil, err
	}
	if candle == nil {
		return nil, nil
	}

	return &analysis.Analytic{
		ID:             uuid.New(),
		Name:           c.Name(),
		Exchange:       indicatorData.Exchange,
		Symbol:         indicatorData.Symbol,
		Unit:           indicatorData.Unit,
		Interval:       indicatorData.Interval,
		Datetime:       indicatorData.Datetime,
		ByIndicator:    indicatorData.Name,
		IndicatorDepth: indicatorData.Depth,
		Depth:          1,
		Value:          float64(compareValues(candle.ClosePrice, indicatorData.Value)),
	}, err
}

func compareValues(left, right float64) int {
	if left > right {
		return 1
	}
	if left < right {
		return -1
	}
	return 0
}

func (c *ratioCandleToMA) candleByIndicator(ctx context.Context, indicator domain.Indicator) (*domain.Candlestick, error) {
	candles, err := c.candleService.SequenceCandlesticksToDate(ctx, indicator.Exchange, indicator.Symbol, string(indicator.Unit), indicator.Interval, 1, indicator.Datetime)
	if err != nil {
		return nil, err
	}
	for _, candle := range candles {
		if candle.StartTime == indicator.Datetime {
			return &candle, nil
		}
	}

	return nil, nil
}
