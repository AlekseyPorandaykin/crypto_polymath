package exchange

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/gateio"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/cenkalti/backoff/v4"
	"strings"
	"time"
)

const GateIoExchange = "gate.io"

type GateIo struct {
	client *gateio.Client
}

func NewGateIo(client *gateio.Client) *GateIo {
	return &GateIo{client: client}
}

func (c *GateIo) Prices(ctx context.Context) ([]price.ExchangeDTO, error) {
	result := make([]price.ExchangeDTO, 0, 5000)
	var ticks []gateio.Tick
	err := backoff.Retry(func() error {
		var err error
		ticks, err = c.client.Ticker(ctx)
		if err != nil {
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, err
	}
	currentTime := time.Now()
	for _, tick := range ticks {
		result = append(result, price.ExchangeDTO{
			Exchange:  GateIoExchange,
			Symbol:    strings.Replace(tick.CurrencyPair, "_", "", 1),
			Value:     tick.LastTradingPrice,
			CreatedAt: currentTime,
		})
	}

	return result, nil
}

func (c *GateIo) Price(ctx context.Context, symbol string) (price.ExchangeDTO, error) {
	return price.ExchangeDTO{}, nil
}
