package listeners

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/service"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/dispatcher"
)

type CreateIndicator struct {
	indicatorHandler *service.IndicatorHandler
}

func NewCreateIndicator(indicatorHandler *service.IndicatorHandler) dispatcher.Listener[domain.CreateIndicatorEventBody] {
	return &CreateIndicator{indicatorHandler: indicatorHandler}
}

func (c *CreateIndicator) Handle(e dispatcher.Event[domain.CreateIndicatorEventBody]) {
	c.indicatorHandler.Calculate(
		context.TODO(),
		e.Body.Exchange,
		e.Body.Symbol,
		e.Body.Unit,
		e.Body.Interval,
	)
}
