package logging

import (
	"context"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/util"
	"go.uber.org/zap"
)

var _ price.Repository = (*PriceRepository)(nil)

type PriceRepository struct {
	inner  price.Repository
	logger *zap.Logger
	db     string
}

func NewPriceRepository(inner price.Repository, logger *zap.Logger, db string) *PriceRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &PriceRepository{inner: inner, logger: logger, db: db}
}

func (r *PriceRepository) Save(ctx context.Context, data ...price.StorageDTO) error {
	defer r.log(ctx, "price_save")()
	return r.inner.Save(ctx, data...)
}

func (r *PriceRepository) Find(ctx context.Context, exchange, symbol string) (*price.StorageDTO, error) {
	defer r.log(ctx, "price_find")()
	return r.inner.Find(ctx, exchange, symbol)
}

func (r *PriceRepository) ListByExchange(ctx context.Context, exchange string) ([]price.StorageDTO, error) {
	defer r.log(ctx, "price_list_by_exchange")()
	return r.inner.ListByExchange(ctx, exchange)
}

func (r *PriceRepository) ListBySymbol(ctx context.Context, symbol string) ([]price.StorageDTO, error) {
	defer r.log(ctx, "price_list_by_symbol")()
	return r.inner.ListBySymbol(ctx, symbol)
}

func (r *PriceRepository) Delete(ctx context.Context, exchange string, to time.Time) error {
	defer r.log(ctx, "price_delete")()
	return r.inner.Delete(ctx, exchange, to)
}

func (r *PriceRepository) log(ctx context.Context, query string) func() {
	now := time.Now()
	return func() {
		r.logger.Debug("Execute query",
			zap.String("query", query),
			zap.String("db", r.db),
			zap.String("duration", time.Since(now).String()),
			zap.String("request_id", util.RequestIDFromContext(ctx)),
		)
	}
}
