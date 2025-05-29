package listeners

import (
	queue_contract "github.com/AlekseyPorandaykin/crypto_polymath/api/queue"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/queue"
	"github.com/AlekseyPorandaykin/go-template/pkg/dispatcher"
	"go.uber.org/zap"
)

type CandleIndicator struct {
	p *queue.RabbitMQProducer[queue_contract.CandleIndicator]
}

func NewCandleIndicator(
	p *queue.RabbitMQProducer[queue_contract.CandleIndicator],
) dispatcher.Listener[candle_indicator.Indicator] {
	return &CandleIndicator{p: p}
}

func (c *CandleIndicator) Handle(e dispatcher.Event[candle_indicator.Indicator]) {
	m := queue_contract.CandleIndicator{
		Name:       e.Body.Name,
		Exchange:   e.Body.Exchange,
		Symbol:     e.Body.Symbol,
		Unit:       string(e.Body.Unit),
		Interval:   e.Body.Interval,
		StartTime:  e.Body.StartTime,
		OpenPrice:  e.Body.OpenPrice,
		HighPrice:  e.Body.HighPrice,
		LowPrice:   e.Body.LowPrice,
		ClosePrice: e.Body.ClosePrice,
	}
	if err := c.p.Publish(m); err != nil {
		zap.L().Error("publish candle_indicator", zap.Error(err), zap.Any("candle_indicator", m))
		return
	}
}
