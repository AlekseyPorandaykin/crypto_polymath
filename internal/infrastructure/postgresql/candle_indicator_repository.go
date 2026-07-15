package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/jmoiron/sqlx"
)

var _ candle_indicator.Repository = (*CandleIndicatorRepository)(nil)

type CandleIndicatorRepository struct {
	db *sqlx.DB
}

func NewCandleIndicatorRepository(db *sqlx.DB) *CandleIndicatorRepository {
	return &CandleIndicatorRepository{db: db}
}
func (repo *CandleIndicatorRepository) Find(
	ctx context.Context, name, exchange, symbol, unit string, interval int, startTime time.Time,
) (*candle_indicator.StorageDTO, error) {
	query := `
SELECT id,
       name,
       exchange,
       symbol,
       unit,
       interval,
       start_time,
       open_price,
       high_price,
       low_price,
       close_price,
       created_at
FROM candlestick_indicators
WHERE name = ?
  AND exchange = ?
  AND symbol = ?
  AND unit = ?
  AND interval = ?
  AND start_time = ?
ORDER BY created_at DESC
LIMIT 1`
	result := candle_indicator.StorageDTO{}
	err := repo.db.GetContext(ctx, &result, repo.db.Rebind(query), name, exchange, symbol, unit, interval, startTime)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (repo *CandleIndicatorRepository) FindMany(
	ctx context.Context, name, exchange, symbol, unit string, interval int, startTimes []time.Time,
) ([]candle_indicator.StorageDTO, error) {
	if len(startTimes) == 0 {
		return nil, nil
	}
	query, args, err := sqlx.In(`
SELECT id,
       name,
       exchange,
       symbol,
       unit,
       interval,
       start_time,
       open_price,
       high_price,
       low_price,
       close_price,
       created_at
FROM candlestick_indicators
WHERE name = ?
  AND exchange = ?
  AND symbol = ?
  AND unit = ?
  AND interval = ?
  AND start_time IN (?)`, name, exchange, symbol, unit, interval, startTimes)
	if err != nil {
		return nil, err
	}
	result := make([]candle_indicator.StorageDTO, 0, len(startTimes))
	if err := repo.db.SelectContext(ctx, &result, repo.db.Rebind(query), args...); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *CandleIndicatorRepository) FetchLast(
	ctx context.Context, name, exchange, symbol, unit string, interval int,
) ([]candle_indicator.StorageDTO, error) {
	query := `
SELECT id,
       name,
       exchange,
       symbol,
       unit,
       interval,
       start_time,
       open_price,
       high_price,
       low_price,
       close_price,
       created_at
FROM candlestick_indicators
WHERE name = ?
  AND exchange = ?
  AND symbol = ?
  AND unit = ?
  AND interval = ?
ORDER BY start_time DESC
LIMIT 100`
	result := make([]candle_indicator.StorageDTO, 0, 100)
	if err := repo.db.SelectContext(ctx, &result, repo.db.Rebind(query), name, exchange, symbol, unit, interval); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *CandleIndicatorRepository) LastAddedFromDate(
	ctx context.Context, name, exchange, unit string, interval int, from time.Time,
) ([]candle_indicator.StorageDTO, error) {
	query := `
SELECT id,
       name,
       exchange,
       symbol,
       unit,
       interval,
       start_time,
       open_price,
       high_price,
       low_price,
       close_price,
       created_at
FROM candlestick_indicators
WHERE name = ?
  AND exchange = ?
  AND unit = ?
  AND interval = ?
	  AND created_at > ?
ORDER BY created_at ASC
LIMIT 100`
	result := make([]candle_indicator.StorageDTO, 0, 100)
	if err := repo.db.SelectContext(ctx, &result, repo.db.Rebind(query), name, exchange, unit, interval, from.In(time.UTC)); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *CandleIndicatorRepository) Save(ctx context.Context, data []candle_indicator.StorageDTO) error {
	if len(data) == 0 {
		return nil
	}
	query := `
INSERT INTO candlestick_indicators (id,
                                    name,
                                    exchange,
                                    symbol,
                                    unit,
                                    interval,
                                    start_time,
                                    open_price,
                                    high_price,
                                    low_price,
                                    close_price,
                                    created_at)
VALUES (:id,
        :name,
        :exchange,
        :symbol,
        :unit,
        :interval,
        :start_time,
        :open_price,
        :high_price,
        :low_price,
        :close_price,
        :created_at)
ON CONFLICT (name, exchange, symbol, unit, interval, start_time) DO NOTHING`
	// BindNamed вызывается один раз на весь срез: sqlx сам размножит VALUES(...) на len(data) групп
	// со сквозной нумерацией плейсхолдеров. При вызове BindNamed по одному элементу на строку
	// нумерация плейсхолдеров ($1..$N) начиналась заново на каждой строке, из-за чего при
	// нескольких строках в одном запросе возникала ошибка "mismatched param and argument count".
	preparedQuery, args, err := repo.db.BindNamed(query, data)
	if err != nil {
		return err
	}
	if _, err := repo.db.ExecContext(ctx, preparedQuery, args...); err != nil {
		return err
	}
	return nil
}
