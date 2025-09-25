package exchange

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_loader/pkg/mexc"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
)

const MexcExchange = "mexc"

type Mexc struct {
	client *mexc.Client
}

func NewMexc(client *mexc.Client) *Mexc {
	return &Mexc{client: client}
}

func (c *Mexc) Prices(ctx context.Context) ([]price.ExchangeDTO, error) {
	result := make([]price.ExchangeDTO, 0, 2000)
	var priceSymbols []mexc.PriceSymbol
	err := backoff.Retry(func() error {
		var err error
		priceSymbols, err = c.client.SymbolPriceTicker(ctx)
		if err != nil {
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, errors.Wrap(err, "error get price from mexc")
	}
	currentTime := time.Now()
	for _, priceSymbol := range priceSymbols {
		result = append(result, price.ExchangeDTO{
			Exchange:  MexcExchange,
			Symbol:    priceSymbol.Symbol,
			Value:     priceSymbol.Price,
			CreatedAt: currentTime,
		})
	}
	return result, nil
}

func (c *Mexc) Price(ctx context.Context, symbol string) (price.ExchangeDTO, error) {
	return price.ExchangeDTO{}, nil
}
