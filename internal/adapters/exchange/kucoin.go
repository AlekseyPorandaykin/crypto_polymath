package exchange

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/kucoin"
	"github.com/AlekseyPorandaykin/crypto_loader/pkg/kucoin/response"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"strings"
	"time"
)

const KucoinExchange = "kucoin"

type Kucoin struct {
	client *kucoin.Client
}

func NewKucoin(client *kucoin.Client) *Kucoin {
	return &Kucoin{
		client: client,
	}
}

func (c *Kucoin) Prices(ctx context.Context) ([]price.ExchangeDTO, error) {
	result := make([]price.ExchangeDTO, 0, 1500)
	var allTickersResp response.AllTickersResponse
	err := backoff.Retry(func() error {
		var err error
		allTickersResp, err = c.client.GetAllTickers(ctx)
		if err != nil {
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, errors.Wrap(err, "error get price from kucoin")
	}
	if !allTickersResp.IsOk() {
		return nil, fmt.Errorf("incorrect response from kucoin: %s", allTickersResp.Code)
	}
	currentTime := time.UnixMilli(allTickersResp.Data.Time)
	if currentTime.Year() != time.Now().Year() {
		currentTime = time.Now()
	}
	for _, item := range allTickersResp.Data.Ticker {
		result = append(result, price.ExchangeDTO{
			Exchange:  KucoinExchange,
			Symbol:    strings.Replace(item.Symbol, "-", "", 1),
			Value:     item.LastPrice,
			CreatedAt: currentTime,
		})
	}
	return result, nil
}

func (c *Kucoin) Price(ctx context.Context, symbol string) (price.ExchangeDTO, error) {
	return price.ExchangeDTO{}, nil
}
