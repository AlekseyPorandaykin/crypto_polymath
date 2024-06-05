package calculators

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/google/uuid"
)

type ratioCandleToEMA struct {
	candleService candlestick.Candlestick
}

func NewRationCandleToEMA(candleService candlestick.Candlestick) analysis.Calculator {
	return &ratioCandleToEMA{candleService: candleService}
}

func (c *ratioCandleToEMA) Name() string {
	return domain.RatioCandleToEMAIndicator
}

func (c *ratioCandleToEMA) ByIndicator() string {
	return domain.EMAIndicator
}

func (c *ratioCandleToEMA) SupportDepth(depth int) bool {
	return depth > 0
}

func (c *ratioCandleToEMA) SupportInterval(interval int) bool {
	return interval > 0
}

func (c *ratioCandleToEMA) Calculate(ctx context.Context, indicatorData domain.Indicator) ([]analysis.Analytic, error) {
	candle, err := c.candleByIndicator(ctx, indicatorData)
	if err != nil {
		return nil, err
	}
	if candle == nil {
		return nil, nil
	}

	return []analysis.Analytic{
		analysis.Analytic{
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
		},
	}, err
}

func (c *ratioCandleToEMA) candleByIndicator(ctx context.Context, indicator domain.Indicator) (*domain.Candlestick, error) {
	candles, err := c.candleService.CandlesticksToDate(ctx, indicator.Exchange, indicator.Symbol, string(indicator.Unit), indicator.Interval, 1, indicator.Datetime)
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
