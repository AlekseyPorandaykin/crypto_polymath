package listeners

import (
	queue_contract "github.com/AlekseyPorandaykin/crypto_polymath/api/queue"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/queue"
	"github.com/AlekseyPorandaykin/go-template/pkg/dispatcher"
	"go.uber.org/zap"
)

type LoadedCandlesticks struct {
	p *queue.RabbitMQProducer[queue_contract.Action]
}

func NewLoadedCandlesticks(
	p *queue.RabbitMQProducer[queue_contract.Action],
) dispatcher.Listener[domain.LoadedCandlesticksActionBody] {
	return &LoadedCandlesticks{p: p}
}

func (c *LoadedCandlesticks) Handle(e dispatcher.Event[domain.LoadedCandlesticksActionBody]) {
	m := queue_contract.Action{
		Name:      domain.LoadedCandlesticksForSymbolAction,
		Exchange:  e.Body.Exchange,
		Symbol:    e.Body.Symbol,
		Unit:      string(e.Body.Unit),
		Interval:  e.Body.Interval,
		CreatedAt: e.Body.CreatedAt,
		Duration:  e.Body.Duration,
	}
	if err := c.p.Publish(m); err != nil {
		zap.L().Error("publish candlestick", zap.Error(err), zap.Any("candlestick", m))
		return
	}
}
