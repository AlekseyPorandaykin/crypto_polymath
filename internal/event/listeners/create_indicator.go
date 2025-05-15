package listeners

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/application"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
	"time"
)

type CreateIndicator struct {
	indicatorHandler *application.IndicatorHandler
}

func NewCreateIndicator(indicatorHandler *application.IndicatorHandler) dispatcher.Listener[domain.CreateIndicatorEventBody] {
	return &CreateIndicator{indicatorHandler: indicatorHandler}
}

func (c *CreateIndicator) Handle(e dispatcher.Event[domain.CreateIndicatorEventBody]) {
	//Может быть, что давно не рассчитывали индикаторы и будет долгий процесс расчета.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()
	c.indicatorHandler.Calculate(
		ctx,
		e.Body.Exchange,
		e.Body.Symbol,
		e.Body.Unit,
		e.Body.Interval,
	)
}
