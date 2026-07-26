package listeners

import (
	queue_contract "github.com/AlekseyPorandaykin/crypto_polymath/api/queue"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/queue"
	"github.com/AlekseyPorandaykin/go-kit/pkg/dispatcher"
	"go.uber.org/zap"
)

type Candlestick struct {
	p queue.Publisher[queue_contract.Candlestick]
}

func NewCandlestick(
	p queue.Publisher[queue_contract.Candlestick],
) dispatcher.Listener[domain.Candlestick] {
	return &Candlestick{p: p}
}

func (c *Candlestick) Handle(e dispatcher.Event[domain.Candlestick]) {
	m := queue_contract.Candlestick{
		Exchange:   e.Body.Exchange,
		Symbol:     e.Body.Symbol,
		Unit:       string(e.Body.Unit),
		Interval:   e.Body.Interval,
		StartTime:  e.Body.StartTime,
		OpenPrice:  e.Body.OpenPrice,
		HighPrice:  e.Body.HighPrice,
		LowPrice:   e.Body.LowPrice,
		ClosePrice: e.Body.ClosePrice,
		Volume:     e.Body.Volume,
	}
	if err := c.p.Publish(m); err != nil {
		zap.L().Error("publish candlestick", zap.Error(err), zap.Any("candlestick", m))
		return
	}
}
