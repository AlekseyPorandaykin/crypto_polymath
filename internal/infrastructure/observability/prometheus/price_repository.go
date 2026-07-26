package prometheus

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/go-kit/pkg/metrics"
)

var _ price.Repository = (*PriceRepository)(nil)

type PriceRepository struct {
	inner price.Repository
	db    string
}

func NewPriceRepository(inner price.Repository, db string) *PriceRepository {
	return &PriceRepository{inner: inner, db: db}
}

func (r *PriceRepository) Save(ctx context.Context, data ...price.StorageDTO) error {
	defer metrics.DBQueryHelper(r.db, "price_save")()
	return r.inner.Save(ctx, data...)
}

func (r *PriceRepository) Find(ctx context.Context, exchange, symbol string) (*price.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "price_find")()
	return r.inner.Find(ctx, exchange, symbol)
}

func (r *PriceRepository) ListByExchange(ctx context.Context, exchange string) ([]price.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "price_list_by_exchange")()
	return r.inner.ListByExchange(ctx, exchange)
}

func (r *PriceRepository) ListBySymbol(ctx context.Context, symbol string) ([]price.StorageDTO, error) {
	defer metrics.DBQueryHelper(r.db, "price_list_by_symbol")()
	return r.inner.ListBySymbol(ctx, symbol)
}

func (r *PriceRepository) Delete(ctx context.Context, exchange string, to time.Time) error {
	defer metrics.DBQueryHelper(r.db, "price_delete")()
	return r.inner.Delete(ctx, exchange, to)
}
