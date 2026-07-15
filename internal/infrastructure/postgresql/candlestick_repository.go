package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/jmoiron/sqlx"
)

var _ candlestick.Repository = (*CandlestickRepository)(nil)

type CandlestickRepository struct {
	db *sqlx.DB
}

func NewCandlestickRepository(db *sqlx.DB) *CandlestickRepository {
	return &CandlestickRepository{db: db}
}

func (repo *CandlestickRepository) Save(ctx context.Context, data ...candlestick.StorageDTO) error {
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
(?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?)
`
	values := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data)*12)

	for _, item := range data {
		values = append(values, valueQuery)
		args = append(
			args,
			[]any{
				item.ID,
				item.Symbol,
				item.Exchange,
				item.Unit,
				item.Interval,
				item.StartTime,
				item.OpenPrice,
				item.HighPrice,
				item.LowPrice,
				item.ClosePrice,
				item.Volume,
				item.CreatedAt,
			}...)
	}
	preparedQuery := fmt.Sprintf("%s %s", query, strings.Join(values, ", "))
	if _, err := repo.db.ExecContext(ctx, repo.db.Rebind(preparedQuery), args...); err != nil {
		return err
	}
	return nil
}

func (repo *CandlestickRepository) Last(ctx context.Context, exchange, symbol, unit string, interval, limit, offset int) ([]candlestick.StorageDTO, error) {
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
	err := repo.db.SelectContext(ctx, &data, repo.db.Rebind(query), exchange, symbol, unit, interval, limit, offset)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}
func (repo *CandlestickRepository) LastToDate(ctx context.Context, exchange, symbol, unit string, interval, limit int, to time.Time) ([]candlestick.StorageDTO, error) {
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
	err := repo.db.SelectContext(ctx, &data, repo.db.Rebind(query), exchange, symbol, unit, interval, to, limit)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *CandlestickRepository) FromDate(ctx context.Context, exchange, symbol, unit string, interval, limit int, to time.Time) ([]candlestick.StorageDTO, error) {
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
	err := repo.db.SelectContext(ctx, &data, repo.db.Rebind(query), exchange, symbol, unit, interval, to, limit)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *CandlestickRepository) DeleteOldRows(ctx context.Context, exchange, symbol, unit string, interval int, to time.Time) error {
	var query = `DELETE FROM candlestick WHERE exchange = ? AND symbol = ? AND unit=? AND interval= ? AND  start_time < ?`
	if _, err := repo.db.ExecContext(ctx, repo.db.Rebind(query), exchange, symbol, unit, interval, to); err != nil {
		return err
	}
	return nil
}
func (repo *CandlestickRepository) DeletePrevRows(ctx context.Context, exchange, symbol, unit string, interval int, to time.Time) error {
	var query = `DELETE FROM candlestick WHERE exchange = ? AND symbol = ? AND unit=? AND interval= ? AND  created_at < ?`
	if _, err := repo.db.ExecContext(ctx, repo.db.Rebind(query), exchange, symbol, unit, interval, to); err != nil {
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

func (repo *CandlestickRepository) AllSymbols(ctx context.Context) ([]string, error) {
	var (
		query = `
SELECT  DISTINCT symbol
FROM candlestick 
ORDER BY  symbol
`
		res []string
	)
	if err := repo.db.SelectContext(ctx, &res, query); err != nil {
		return nil, err
	}

	return res, nil
}
