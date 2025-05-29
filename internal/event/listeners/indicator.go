package listeners

import (
	queue_contract "github.com/AlekseyPorandaykin/crypto_polymath/api/queue"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/queue"
	"github.com/AlekseyPorandaykin/go-template/pkg/dispatcher"
	"go.uber.org/zap"
)

type Indicator struct {
	p *queue.RabbitMQProducer[queue_contract.Indicator]
}

func NewIndicator(
	p *queue.RabbitMQProducer[queue_contract.Indicator],
) dispatcher.Listener[domain.Indicator] {
	return &Indicator{p: p}
}

func (c *Indicator) Handle(e dispatcher.Event[domain.Indicator]) {
	m := queue_contract.Indicator{
		Exchange: e.Body.Exchange,
		Symbol:   e.Body.Symbol,
		Unit:     string(e.Body.Unit),
		Interval: e.Body.Interval,
		Datetime: e.Body.Datetime,
		Name:     e.Body.Name,
		Depth:    e.Body.Depth,
		Value:    e.Body.Value,
	}
	if err := c.p.Publish(m); err != nil {
		zap.L().Error("publish indicator", zap.Error(err), zap.Any("indicator", m))
		return
	}
}
