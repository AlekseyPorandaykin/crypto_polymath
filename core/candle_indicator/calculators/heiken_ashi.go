package calculators

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/util"
	"github.com/pkg/errors"
	"math"
)

type heikenAshi struct {
	repo candle_indicator.Repository
}

func NewHeikenAshi(repo candle_indicator.Repository) candle_indicator.Calculator {
	return &heikenAshi{repo: repo}
}

func (c *heikenAshi) Name() string {
	return domain.HeikenAshiIndicator

}

func (c *heikenAshi) Calculate(ctx context.Context, candle domain.Candlestick) (*candle_indicator.Indicator, error) {
	prevStorageIndicator, err := c.repo.Find(ctx, c.Name(), candle.Exchange, candle.Symbol, string(candle.Unit), candle.Interval, candle.PrevStartTime())
	if err != nil {
		return nil, errors.Wrap(err, "get prev ha candle")
	}
	openPriceHA := candle.OpenPrice
	if prevStorageIndicator != nil {
		prevIndicatorHA := candle_indicator.StorageToDomain(*prevStorageIndicator)
		//предыдущий хейкен айши
		openPriceHA = (prevIndicatorHA.ClosePrice + prevIndicatorHA.OpenPrice) / 2
	}
	closePriceHA := (candle.OpenPrice + candle.ClosePrice + candle.HighPrice + candle.LowPrice) / 4
	return &candle_indicator.Indicator{
		Name:       domain.HeikenAshiIndicator,
		Exchange:   candle.Exchange,
		Symbol:     candle.Symbol,
		Unit:       candle.Unit,
		Interval:   candle.Interval,
		StartTime:  candle.StartTime,
		OpenPrice:  util.RoundCoin(openPriceHA, 6),
		HighPrice:  util.RoundCoin(math.Max(candle.HighPrice, math.Max(openPriceHA, closePriceHA)), 6),
		LowPrice:   util.RoundCoin(math.Min(candle.LowPrice, math.Min(openPriceHA, closePriceHA)), 6),
		ClosePrice: util.RoundCoin(closePriceHA, 6),
	}, nil
}
