package repository

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/pkg/errors"
)

var _ price.Repository = (*PriceRepository)(nil)

type PriceRepository struct {
	storage price.Repository
	cache   price.Repository
}

func NewPriceRepository(storage, cache price.Repository) price.Repository {
	return &PriceRepository{cache: cache, storage: storage}
}

func (p *PriceRepository) Save(ctx context.Context, data ...price.StorageDTO) error {
	_ = p.cache.Save(ctx, data...)
	if err := p.storage.Save(ctx, data...); err != nil {
		return errors.Wrap(err, "save to storage")
	}
	return nil
}

func (p *PriceRepository) Find(ctx context.Context, exchange, symbol string) (*price.StorageDTO, error) {
	dataCache, err := p.cache.Find(ctx, exchange, symbol)
	if err != nil {
		return nil, errors.Wrap(err, "from cache")
	}
	if dataCache != nil {
		return dataCache, nil
	}
	dataStorage, err := p.storage.Find(ctx, exchange, symbol)
	if err != nil {
		return nil, errors.Wrap(err, "from storage")
	}
	return dataStorage, nil
}

func (p *PriceRepository) ListByExchange(ctx context.Context, exchange string) ([]price.StorageDTO, error) {
	dataCache, err := p.cache.ListByExchange(ctx, exchange)
	if err != nil {
		return nil, errors.Wrap(err, "from cache")
	}
	if dataCache != nil {
		return dataCache, nil
	}
	dataStorage, err := p.storage.ListByExchange(ctx, exchange)
	if err != nil {
		return nil, errors.Wrap(err, "from storage")
	}
	return dataStorage, nil
}

func (p *PriceRepository) ListBySymbol(ctx context.Context, symbol string) ([]price.StorageDTO, error) {
	dataCache, err := p.cache.ListBySymbol(ctx, symbol)
	if err != nil {
		return nil, errors.Wrap(err, "from cache")
	}
	if dataCache != nil {
		return dataCache, nil
	}
	dataStorage, err := p.storage.ListBySymbol(ctx, symbol)
	if err != nil {
		return nil, errors.Wrap(err, "from storage")
	}
	return dataStorage, nil
}

func (p *PriceRepository) Delete(ctx context.Context, exchange string, to time.Time) error {
	_ = p.cache.Delete(ctx, exchange, to)
	return p.storage.Delete(ctx, exchange, to)
}
