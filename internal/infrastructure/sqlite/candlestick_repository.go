package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/metrics"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

var _ candlestick.Repository = (*CandlestickRepository)(nil)

type CandlestickRepository struct {
	db *sqlx.DB
}

func NewCandlestickRepository(db *sqlx.DB) *CandlestickRepository {
	return &CandlestickRepository{db: db}
}

func (repo *CandlestickRepository) SaveBatch(ctx context.Context, data ...candlestick.StorageDTO) error {
	defer metrics.DBQueryHelper("crypto_polymath", "candlestick_save")()
	query := `
INSERT INTO candlestick(id,
                        symbol,
                        exchange,
                        unit,
                        interval,
                        start_time,
                        open_price,
                        high_price,
                        low_price,
                        close_price,
                        volume,
                        created_at)
VALUES 
`
	valueQuery := `
(:id,
        :symbol,
        :exchange,
        :unit,
        :interval,
        :start_time,
        :open_price,
        :high_price,
        :low_price,
        :close_price,
        :volume,
        :created_at)
`
	values := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data)*12)

	for _, item := range data {
		preparedValues, argsItem, err := repo.db.BindNamed(valueQuery, item)
		if err != nil {
			return nil
		}
		values = append(values, preparedValues)
		args = append(args, argsItem...)
	}
	preparedQuery := fmt.Sprintf("%s %s", query, strings.Join(values, ", "))
	if _, err := repo.db.ExecContext(ctx, preparedQuery, args...); err != nil {
		return err
	}
	return nil
}

func (repo *CandlestickRepository) Last(ctx context.Context, exchange, symbol, unit string, interval, limit, offset int) ([]candlestick.StorageDTO, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "candlestick_last")()
	var (
		query = `
SELECT id,
       symbol,
       exchange,
       unit,
       interval,
       start_time,
       open_price,
       high_price,
       low_price,
       close_price,
       volume,
       created_at
FROM candlestick
WHERE exchange =? AND symbol = ? AND unit = ? AND interval = ?
ORDER BY start_time DESC
LIMIT ? OFFSET ?
`
		data = make([]candlestick.StorageDTO, 0, limit)
	)
	err := repo.db.SelectContext(ctx, &data, query, exchange, symbol, unit, interval, limit, offset)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}
func (repo *CandlestickRepository) LastToDate(ctx context.Context, exchange, symbol, unit string, interval, limit int, to time.Time) ([]candlestick.StorageDTO, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "candlestick_last_to_date")()
	var (
		query = `
SELECT id,
       symbol,
       exchange,
       unit,
       interval,
       start_time,
       open_price,
       high_price,
       low_price,
       close_price,
       volume,
       created_at
FROM candlestick
WHERE exchange =? AND symbol = ? AND unit = ? AND interval = ? AND start_time <= ?
ORDER BY start_time DESC
LIMIT ?
`
		data = make([]candlestick.StorageDTO, 0, limit)
	)
	err := repo.db.SelectContext(ctx, &data, query, exchange, symbol, unit, interval, to, limit)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *CandlestickRepository) FromDate(ctx context.Context, exchange, symbol, unit string, interval, limit int, to time.Time) ([]candlestick.StorageDTO, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "candlestick_find_from")()
	var (
		query = `
SELECT id,
       symbol,
       exchange,
       unit,
       interval,
       start_time,
       open_price,
       high_price,
       low_price,
       close_price,
       volume,
       created_at
FROM candlestick
WHERE exchange =? AND symbol = ? AND unit = ? AND interval = ? AND start_time >= ?
ORDER BY start_time ASC
LIMIT ?
`
		data = make([]candlestick.StorageDTO, 0, limit)
	)
	err := repo.db.SelectContext(ctx, &data, query, exchange, symbol, unit, interval, to, limit)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *CandlestickRepository) DeleteOldRows(ctx context.Context, exchange, symbol, unit string, interval int, to time.Time) error {
	defer metrics.DBQueryHelper("crypto_polymath", "candlestick_delete_old_rows")()
	var query = `DELETE FROM candlestick WHERE exchange = ? AND symbol = ? AND unit=? AND interval= ? AND  start_time < ?`
	if _, err := repo.db.ExecContext(ctx, query, exchange, symbol, unit, interval, to); err != nil {
		return err
	}
	return nil
}

func (repo *CandlestickRepository) ListUniq(ctx context.Context) ([]candlestick.UniqDTO, error) {
	var (
		query = `
SELECT  exchange,symbol, unit, interval
FROM candlestick
GROUP BY exchange,symbol,  unit, interval
ORDER BY  exchange,symbol, unit, interval
`
		res []candlestick.UniqDTO
	)
	if err := repo.db.SelectContext(ctx, &res, query); err != nil {
		return nil, err
	}

	return res, nil
}
