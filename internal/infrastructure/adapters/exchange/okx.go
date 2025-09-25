package exchange

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AlekseyPorandaykin/crypto_loader/pkg/okx"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/okx/response"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
)

const OkxExchange = "okx"

type Okx struct {
	client *okx.Client
}

func NewOkx(client *okx.Client) *Okx {
	return &Okx{client: client}
}

func (c *Okx) Prices(ctx context.Context) ([]price.ExchangeDTO, error) {
	result := make([]price.ExchangeDTO, 0, 500)
	var tickerResp response.TickersResponse
	err := backoff.Retry(func() error {
		var err error
		tickerResp, err = c.client.Tickers(ctx)
		if err != nil {
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, errors.Wrap(err, "get price from okx")
	}
	if !tickerResp.IsOk() {
		return nil, fmt.Errorf("incorrect response from okx: %s", tickerResp.Message)
	}
	currentTime := time.Now()
	for _, item := range tickerResp.Data {
		result = append(result, price.ExchangeDTO{
			Exchange:  OkxExchange,
			Symbol:    strings.Replace(item.InstrumentID, "-", "", 1),
			Value:     item.LastTradedPrice,
			CreatedAt: currentTime,
		})
	}
	return result, nil
}

func (c *Okx) Price(ctx context.Context, symbol string) (price.ExchangeDTO, error) {
	return price.ExchangeDTO{}, nil
}
