package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/metrics"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
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
	defer metrics.DBQueryHelper("crypto_polymath", "candle_indicator_find")()
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
	err := repo.db.GetContext(ctx, &result, query, name, exchange, symbol, unit, interval, startTime)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (repo *CandleIndicatorRepository) FetchLast(
	ctx context.Context, name, exchange, symbol, unit string, interval int,
) ([]candle_indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "candle_indicator_fetch_last")()
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
	if err := repo.db.SelectContext(ctx, &result, query, name, exchange, symbol, unit, interval); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *CandleIndicatorRepository) LastAddedFromDate(
	ctx context.Context, name, exchange, unit string, interval int, from time.Time,
) ([]candle_indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "candle_indicator_fetch_last")()
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
	if err := repo.db.SelectContext(ctx, &result, query, name, exchange, unit, interval, from.In(time.UTC)); err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *CandleIndicatorRepository) Save(ctx context.Context, data []candle_indicator.StorageDTO) error {
	if len(data) == 0 {
		return nil
	}
	defer metrics.DBQueryHelper("crypto_polymath", "candle_indicator_save")()
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
VALUES 
`
	valueQuery := `
(:id,
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
