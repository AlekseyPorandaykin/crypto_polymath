package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/price"
	"github.com/jmoiron/sqlx"
)

var _ price.Repository = (*PriceRepository)(nil)

type PriceRepository struct {
	db *sqlx.DB
}

func NewPriceRepository(db *sqlx.DB) *PriceRepository {
	return &PriceRepository{db: db}
}

func (p *PriceRepository) Save(ctx context.Context, data ...price.StorageDTO) error {
	const query = `
INSERT INTO 
    prices(id, symbol, exchange, value, created_at) VALUES 
`
	var queryValues = `(?, ?, ?, ?, ?)`
	values := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data)*5)
	for _, item := range data {
		if math.IsNaN(item.Value) {
			continue
		}
		values = append(values, queryValues)
		args = append(args, []any{item.ID, item.Symbol, item.Exchange, item.Value, item.CreatedAt}...)
	}
	preparedQuery := query + strings.Join(values, ", ")
	if _, err := p.db.ExecContext(ctx, p.db.Rebind(preparedQuery), args...); err != nil {
		return err
	}
	return nil
}

func (p *PriceRepository) Find(ctx context.Context, exchange, symbol string) (*price.StorageDTO, error) {
	var (
		query = `
SELECT id, symbol, exchange, value, created_at
FROM prices
WHERE exchange = ?
  AND symbol = ?
ORDER BY created_at DESC
LIMIT 1
`
		res price.StorageDTO
	)
	err := p.db.GetContext(ctx, &res, p.db.Rebind(query), exchange, symbol)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (p *PriceRepository) ListByExchange(ctx context.Context, exchange string) ([]price.StorageDTO, error) {
	var (
		query = `
SELECT id, symbol, exchange, value, created_at
FROM prices
WHERE exchange = ?
ORDER BY created_at DESC
`
		result []price.StorageDTO
	)
	err := p.db.SelectContext(ctx, &result, p.db.Rebind(query), exchange)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (p *PriceRepository) ListBySymbol(ctx context.Context, symbol string) ([]price.StorageDTO, error) {
	var (
		query = `
SELECT id, symbol, exchange, value, created_at
FROM prices
WHERE symbol = ?
ORDER BY created_at DESC
`
		result []price.StorageDTO
	)
	err := p.db.SelectContext(ctx, &result, p.db.Rebind(query), symbol)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *PriceRepository) Delete(ctx context.Context, exchange string, to time.Time) error {
	var query = `
DELETE FROM prices WHERE exchange=? AND created_at < ?
`
	if _, err := p.db.ExecContext(ctx, p.db.Rebind(query), exchange, to); err != nil {
		return err
	}
	return nil
}
