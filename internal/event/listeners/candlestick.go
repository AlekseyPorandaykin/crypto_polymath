package listeners

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/service"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
)

type Candlestick struct {
	indicatorHandler *service.IndicatorHandler
}

func NewCandlestick(indicatorHandler *service.IndicatorHandler) dispatcher.Listener[domain.Candlestick] {
	return &Candlestick{indicatorHandler: indicatorHandler}
}

func (c *Candlestick) Handle(e dispatcher.Event[domain.Candlestick]) {
	c.indicatorHandler.CalculateByCandle(context.TODO(), e.Body)
}
