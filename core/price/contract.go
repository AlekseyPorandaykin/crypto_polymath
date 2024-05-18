package price

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/google/uuid"
	"time"
)

type ExchangeDTO struct {
	Symbol    string
	Exchange  string
	Value     string
	CreatedAt time.Time
}

type ExchangeLoader interface {
	Prices(ctx context.Context) ([]ExchangeDTO, error)
	Price(ctx context.Context, symbol string) (ExchangeDTO, error)
}

type Price interface {
	AddLoader(exchange string, loader ExchangeLoader)
	LoadPrices(ctx context.Context, exchange string) ([]domain.Price, error)
	Save(ctx context.Context, data ...domain.Price) error
	LastPrice(ctx context.Context, exchange, symbol string) (*domain.Price, error)
	LastPricesByExchange(ctx context.Context, exchange string) ([]domain.Price, error)
	LastPricesBySymbol(ctx context.Context, symbol string) ([]domain.Price, error)
	DeleteOldRaws(ctx context.Context, exchange string, to time.Time) error
}

func toStorageDTO(data domain.Price) StorageDTO {
	return StorageDTO{
		ID:        uuid.New(),
		Symbol:    data.Symbol,
		Exchange:  data.Exchange,
		Value:     data.Value,
		CreatedAt: time.Now().In(time.UTC),
	}
}

func fromStorageDTO(data StorageDTO) domain.Price {
	return domain.Price{
		Symbol:   data.Symbol,
		Exchange: data.Exchange,
		Value:    data.Value,
	}
}
