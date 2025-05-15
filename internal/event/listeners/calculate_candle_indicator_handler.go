package listeners

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
	"go.uber.org/zap"
	"time"
)

type CalculateCandleIndicatorsHandler struct {
	candleIndicator candle_indicator.CandleIndicator
}

func NewCalculateCandleIndicatorsHandler(
	candleIndicator candle_indicator.CandleIndicator,
) *CalculateCandleIndicatorsHandler {
	return &CalculateCandleIndicatorsHandler{candleIndicator: candleIndicator}
}

func (l *CalculateCandleIndicatorsHandler) ReceiveNewEvents() {

}

func (l *CalculateCandleIndicatorsHandler) Handle(e dispatcher.Event[domain.ActionBody]) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	if e.Name != domain.LoadedCandlesticksForSymbolAction {
		return
	}
	_, err := l.candleIndicator.CalculateAllIndicators(ctx, e.Body.Exchange, e.Body.Symbol, e.Body.Unit, e.Body.Interval)
	if err != nil {
		zap.L().Error("", zap.Error(err), zap.Any("body", e.Body))
	}
}
