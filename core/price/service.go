package price

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

type service struct {
	loaders map[string]ExchangeLoader
	repo    Repository
}

func NewService(repo Repository) Price {
	return &service{
		loaders: make(map[string]ExchangeLoader),
		repo:    repo,
	}
}

func (s *service) AddLoader(exchange string, loader ExchangeLoader) {
	s.loaders[exchange] = loader
}

func (s *service) LastPricesByExchange(ctx context.Context, exchange string) ([]domain.Price, error) {
	data, err := s.repo.ListByExchange(ctx, exchange)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("get prices for exchange=%s", exchange))
	}
	prices := make([]domain.Price, 0, len(data))
	for _, item := range data {
		prices = append(prices, fromStorageDTO(item))
	}
	return prices, nil
}
func (s *service) LastPricesBySymbol(ctx context.Context, symbol string) ([]domain.Price, error) {
	data, err := s.repo.ListBySymbol(ctx, symbol)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("get prices for symbol=%s", symbol))
	}
	prices := make([]domain.Price, 0, len(data))
	for _, item := range data {
		prices = append(prices, fromStorageDTO(item))
	}
	return prices, nil
}

func (s *service) LastPrice(ctx context.Context, exchange, symbol string) (*domain.Price, error) {
	data, err := s.repo.Find(ctx, exchange, symbol)
	if err != nil {
		return nil, errors.Wrap(err, "find price")
	}
	if data == nil {
		return nil, nil
	}
	p := fromStorageDTO(*data)
	return &p, nil
}

func (s *service) LoadPrices(ctx context.Context, exchange string) ([]domain.Price, error) {
	now := time.Now().In(time.UTC)
	data, err := s.pricesFromExchange(ctx, exchange)
	if err != nil {
		return nil, err
	}
	if err := s.Save(ctx, data...); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Save exchange (%s)", exchange))
	}
	if err := s.repo.Delete(ctx, exchange, now); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("delete old prices (%s)", exchange))
	}
	return data, nil
}

func (s *service) DeleteOldRaws(ctx context.Context, exchange string, to time.Time) error {
	if err := s.repo.Delete(ctx, exchange, to); err != nil {
		return errors.Wrap(err, "delete old raws")
	}
	return nil
}

func (s *service) priceFromExchange(ctx context.Context, exchange, symbol string) (*domain.Price, error) {
	loader := s.loaders[exchange]
	if loader == nil {
		return nil, nil
	}

	data, err := loader.Price(ctx, symbol)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("load symbol(%s) price from exchange (%s)", symbol, exchange))
	}
	val, errPrice := strconv.ParseFloat(data.Value, 64)
	if errPrice != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("parse price value(%s)", data.Value))
	}
	return &domain.Price{
		Symbol:   data.Symbol,
		Exchange: exchange,
		Value:    val,
	}, nil
}

func (s *service) pricesFromExchange(ctx context.Context, exchange string) ([]domain.Price, error) {
	loader := s.loaders[exchange]
	if loader == nil {
		return nil, nil
	}
	data, errLoad := loader.Prices(ctx)
	if errLoad != nil {
		return nil, errors.Wrap(errLoad, "load prices from exchange")
	}
	prices := make([]domain.Price, 0, len(data))
	for _, item := range data {
		if strings.TrimSpace(item.Value) == "" {
			continue
		}
		val, errPrice := strconv.ParseFloat(item.Value, 64)
		if errPrice != nil {
			return nil, errors.Wrap(errPrice, fmt.Sprintf("parse price value(%s)", item.Value))
		}
		prices = append(prices, domain.Price{
			Symbol:   item.Symbol,
			Exchange: exchange,
			Value:    val,
		})
	}
	return prices, nil
}

func (s *service) Save(ctx context.Context, data ...domain.Price) error {
	batch := make([]StorageDTO, 0, len(data))
	for _, item := range data {
		batch = append(batch, toStorageDTO(item))
	}
	err := backoff.Retry(func() error {
		return s.repo.Save(ctx, batch...)
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return errors.Wrap(err, "Save price to storage")
	}
	return nil
}

func (s *service) deleteOldPrices(ctx context.Context, exchange string, to time.Time) error {
	if err := s.repo.Delete(ctx, exchange, to); err != nil {
		return err
	}
	return nil
}
