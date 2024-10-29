package listeners

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/application"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
	"time"
)

type Candlestick struct {
	indicatorHandler *application.IndicatorHandler
}

func NewCandlestick(indicatorHandler *application.IndicatorHandler) dispatcher.Listener[domain.Candlestick] {
	return &Candlestick{indicatorHandler: indicatorHandler}
}

func (c *Candlestick) Handle(e dispatcher.Event[domain.Candlestick]) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	c.indicatorHandler.CalculateByCandle(ctx, e.Body)
}
