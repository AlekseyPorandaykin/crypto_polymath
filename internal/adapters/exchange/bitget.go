package exchange

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/bitget"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"time"
)

const BitgetExchange = "bitget"

type Bitget struct {
	client *bitget.Client
}

func NewBitget(client *bitget.Client) *Bitget {
	return &Bitget{client: client}
}

func (c *Bitget) Prices(ctx context.Context) ([]price.ExchangeDTO, error) {
	result := make([]price.ExchangeDTO, 0, 500)
	var tickerResp bitget.TickersResponse
	err := backoff.Retry(func() error {
		var err error
		tickerResp, err = c.client.GetTicker(ctx)
		if err != nil {
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, errors.Wrap(err, "error get price from bitget")
	}
	if !tickerResp.IsOk() {
		return nil, fmt.Errorf("incorrect response from bitget: %s", tickerResp.Message)
	}
	currentTime := time.UnixMilli(tickerResp.RequestTime)
	if currentTime.Year() != time.Now().Year() {
		currentTime = time.Now()
	}
	for _, tick := range tickerResp.Data {
		result = append(result, price.ExchangeDTO{
			Exchange:  BitgetExchange,
			Symbol:    tick.Symbol,
			Value:     tick.LastPrice,
			CreatedAt: currentTime,
		})
	}
	return result, nil
}

func (c *Bitget) Price(ctx context.Context, symbol string) (price.ExchangeDTO, error) {
	return price.ExchangeDTO{}, nil
}
