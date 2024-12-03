package calculators

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/pkg/errors"
	"math"
)

type heikenAshi struct {
	candlestickService candlestick.Candlestick
}

func NewHeikenAshi(candlestickService candlestick.Candlestick) candle_indicator.Calculator {
	return &heikenAshi{candlestickService: candlestickService}
}

func (c *heikenAshi) Name() string {
	return domain.HeikenAshiIndicator

}

func (c *heikenAshi) Calculate(ctx context.Context, candle domain.Candlestick) (*candle_indicator.Indicator, error) {
	prevCandles, err := c.candlestickService.SequenceCandlesticksToDate(
		ctx,
		candle.Exchange,
		candle.Symbol,
		string(candle.Unit),
		candle.Interval,
		2,
		candle.StartTime,
	)
	if err != nil {
		return nil, errors.Wrap(err, "get prev candle")
	}
	if len(prevCandles) < 2 {
		return nil, nil
	}
	openPrice := (prevCandles[1].ClosePrice + prevCandles[1].OpenPrice) / 2
	return &candle_indicator.Indicator{
		Name:       domain.HeikenAshiIndicator,
		Exchange:   candle.Exchange,
		Symbol:     candle.Symbol,
		Unit:       candle.Unit,
		Interval:   candle.Interval,
		StartTime:  candle.StartTime,
		OpenPrice:  openPrice,
		HighPrice:  math.Max(candle.HighPrice, openPrice),
		LowPrice:   math.Min(candle.HighPrice, openPrice),
		ClosePrice: candle.ClosePrice,
	}, nil
}
