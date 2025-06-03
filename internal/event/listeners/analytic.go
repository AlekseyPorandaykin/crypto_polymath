package listeners

import (
	queue_contract "github.com/AlekseyPorandaykin/crypto_polymath/api/queue"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/queue"
	"github.com/AlekseyPorandaykin/go-template/pkg/dispatcher"
	"go.uber.org/zap"
)

type Analytic struct {
	p queue.Publisher[queue_contract.Analytic]
}

func NewAnalytic(
	p queue.Publisher[queue_contract.Analytic],
) dispatcher.Listener[analysis.Analytic] {
	return &Analytic{p: p}
}

func (c *Analytic) Handle(e dispatcher.Event[analysis.Analytic]) {
	m := queue_contract.Analytic{
		ID:             e.Body.ID,
		Exchange:       e.Body.Exchange,
		Symbol:         e.Body.Symbol,
		Unit:           string(e.Body.Unit),
		Interval:       e.Body.Interval,
		Name:           e.Body.Name,
		Datetime:       e.Body.Datetime,
		Depth:          e.Body.Depth,
		ByIndicator:    e.Body.ByIndicator,
		IndicatorDepth: e.Body.IndicatorDepth,
		Value:          e.Body.Value,
	}
	if err := c.p.Publish(m); err != nil {
		zap.L().Error("publish analytic", zap.Error(err), zap.Any("analytic", m))
		return
	}
}
