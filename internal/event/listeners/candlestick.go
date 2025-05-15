package listeners

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/application"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
	"go.uber.org/zap"
	"time"
)

type Candlestick struct {
	indicatorHandler *application.IndicatorHandler
	candleIndicator  candle_indicator.CandleIndicator
}

func NewCandlestick(indicatorHandler *application.IndicatorHandler, candleIndicator candle_indicator.CandleIndicator) dispatcher.Listener[domain.Candlestick] {
	return &Candlestick{indicatorHandler: indicatorHandler, candleIndicator: candleIndicator}
}

func (c *Candlestick) Handle(e dispatcher.Event[domain.Candlestick]) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	c.indicatorHandler.CalculateByCandle(ctx, e.Body)
	data := []domain.Candlestick{e.Body}
	if _, err := c.candleIndicator.CalculateFromCandlesticks(ctx, data); err != nil {
		zap.L().Error("calculate candle indicator", zap.Error(err), zap.Any("candle", e.Body))
	}
}
