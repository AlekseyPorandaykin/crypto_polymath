package exchange

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_loader/pkg/binance"
	binance_domain "github.com/AlekseyPorandaykin/crypto_loader/pkg/binance/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
)

const BinanceExchange = "binance"

type Binance struct {
	client *binance.Manager
}

func NewBinance(client *binance.Manager) *Binance {
	return &Binance{client: client}
}

func (c *Binance) Prices(ctx context.Context) ([]price.ExchangeDTO, error) {
	result := make([]price.ExchangeDTO, 0, 2500)
	var binancePrices []binance_domain.PriceSymbolDTO
	err := backoff.Retry(func() error {
		var err error
		binancePrices, err = c.client.GetPrice(ctx)
		if err != nil {
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, errors.Wrap(err, "error get price from binance")
	}
	now := time.Now()
	for _, binancePrice := range binancePrices {
		result = append(result, price.ExchangeDTO{
			Exchange:  BinanceExchange,
			Symbol:    binancePrice.Symbol,
			Value:     binancePrice.Price,
			CreatedAt: now,
		})
	}
	return result, nil
}

func (c *Binance) Price(ctx context.Context, symbol string) (price.ExchangeDTO, error) {
	return price.ExchangeDTO{}, nil
}
