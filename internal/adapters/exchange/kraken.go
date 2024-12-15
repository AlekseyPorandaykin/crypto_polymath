package exchange

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/kraken"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"strings"
	"time"
)

const KrakenExchange = "kraken"

type Kraken struct {
	client *kraken.Client
}

func NewKraken(client *kraken.Client) *Kraken {
	return &Kraken{client: client}
}

func (c *Kraken) Prices(ctx context.Context) ([]price.ExchangeDTO, error) {
	result := make([]price.ExchangeDTO, 0, 2500)
	var response kraken.TickerResponse
	err := backoff.Retry(func() error {
		var err error
		response, err = c.client.Ticker(ctx)
		if err != nil {
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, errors.Wrap(err, "error get price from kraken")
	}
	currentTime := time.Now()
	for symbolPair, tick := range response.Result {
		averagePrice, err := tick.AveragePrice()
		if err != nil {
			return nil, errors.Wrap(err, "error get average price from kraken")
		}
		result = append(result, price.ExchangeDTO{
			Exchange:  KrakenExchange,
			Symbol:    strings.Replace(string(symbolPair), "XBT", "BTC", 1),
			Value:     averagePrice,
			CreatedAt: currentTime,
		})
	}
	return result, nil
}

func (c *Kraken) Price(ctx context.Context, symbol string) (price.ExchangeDTO, error) {
	return price.ExchangeDTO{}, nil
}
