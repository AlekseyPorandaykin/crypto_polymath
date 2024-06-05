package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/metrics"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

var _ indicator.Repository = (*IndicatorRepository)(nil)

type IndicatorRepository struct {
	db *sqlx.DB
}

func NewIndicatorRepository(db *sqlx.DB) *IndicatorRepository {
	return &IndicatorRepository{db: db}
}

func (repo *IndicatorRepository) Save(ctx context.Context, data ...indicator.StorageDTO) error {
	defer metrics.DBQueryHelper("crypto_polymath", "indicators_save")()
	var query = `
INSERT INTO indicators(id,
                       symbol,
                       exchange,
                       unit,
                       interval,
                       name,
                       datetime,
                       depth,
                       value,
                       created_at)
VALUES 
`
	valueQuery := `(:id,:symbol,:exchange,:unit,:interval,:name,:datetime,:depth,:value,:created_at)`
	preparedValueQueries := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data)*10)
	for _, item := range data {
		preparedValueQuery, argVals, err := repo.db.BindNamed(valueQuery, item)
		if err != nil {
			return err
		}
		preparedValueQueries = append(preparedValueQueries, preparedValueQuery)
		args = append(args, argVals...)
	}
	preparedQuery := fmt.Sprintf("%s %s", query, strings.Join(preparedValueQueries, ", "))
	if _, err := repo.db.ExecContext(ctx, preparedQuery, args...); err != nil {
		return err
	}
	return nil
}

func (repo *IndicatorRepository) Find(
	ctx context.Context, exchange, symbol, unit string, interval int, datetime time.Time, name string, depth int,
) (*indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "indicators_find")()
	var (
		query = `
SELECT id,
       symbol,
       exchange,
       unit,
       interval,
       name,
       datetime,
       depth,
       value,
       created_at
FROM indicators
WHERE exchange = ? AND symbol = ? AND unit =? AND interval = ? AND datetime = ? AND name = ? AND depth = ?
ORDER BY created_at DESC 
LIMIT 1
`
		res indicator.StorageDTO
	)
	err := repo.db.GetContext(ctx, &res, query, exchange, symbol, unit, interval, datetime, name, depth)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &res, nil
}
func (repo *IndicatorRepository) List(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth, limit, offset int,
) ([]indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "indicators_list")()
	var (
		query = `
SELECT id,
       symbol,
       exchange,
       unit,
       interval,
       name,
       datetime,
       depth,
       value,
       created_at
FROM indicators
WHERE exchange = ? AND symbol = ? AND unit =? AND interval = ? AND name = ? AND depth = ?
ORDER BY created_at DESC 
LIMIT ? OFFSET ?
`
		res []indicator.StorageDTO
	)
	err := repo.db.SelectContext(ctx, &res, query, exchange, symbol, unit, interval, name, depth, limit, offset)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return res, nil
}
func (repo *IndicatorRepository) LastToDate(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth, limit int, to time.Time,
) ([]indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "indicators_list")()
	var (
		query = `
SELECT id,
       symbol,
       exchange,
       unit,
       interval,
       name,
       datetime,
       depth,
       value,
       created_at
FROM indicators
WHERE exchange = ? AND symbol = ? AND unit =? AND interval = ? AND name = ? AND depth = ? AND datetime <= ?
ORDER BY datetime DESC 
LIMIT ? 
`
		res = make([]indicator.StorageDTO, 0, limit)
	)
	err := repo.db.SelectContext(ctx, &res, query, exchange, symbol, unit, interval, name, depth, to, limit)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return res, nil
}
func (repo *IndicatorRepository) Last(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth int,
) (*indicator.StorageDTO, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "indicators_last")()
	var (
		query = `
SELECT id,
       symbol,
       exchange,
       unit,
       interval,
       name,
       datetime,
       depth,
       value,
       created_at
FROM indicators
WHERE exchange = ? AND symbol = ? AND unit =? AND interval = ? AND name = ? AND depth = ?
ORDER BY created_at DESC 
LIMIT 1
`
		res indicator.StorageDTO
	)
	err := repo.db.GetContext(ctx, &res, query, exchange, symbol, unit, interval, name, depth)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &res, nil
}

func (repo *IndicatorRepository) ListUniq(ctx context.Context) ([]indicator.UniqDTO, error) {
	defer metrics.DBQueryHelper("crypto_polymath", "indicators_list_uniq")()
	var (
		query = `
SELECT symbol,
       exchange,
       unit,
       interval,
       name,
       depth
FROM indicators
GROUP BY symbol, exchange, unit, interval, name, depth
ORDER BY symbol, exchange, unit, interval
`
		res []indicator.UniqDTO
	)
	err := repo.db.SelectContext(ctx, &res, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return res, nil
}

func (repo *IndicatorRepository) DeleteOldRows(
	ctx context.Context, symbol, exchangeName, unit string, interval int, name string, depth int, to time.Time,
) error {
	defer metrics.DBQueryHelper("crypto_polymath", "indicators_delete_old_rows")()
	var query = `
DELETE
FROM indicators
WHERE  symbol = ?
	AND exchange = ?
	AND unit = ?
	AND interval = ?
	AND name = ?
	AND depth = ? 
	AND datetime <?
`
	if _, err := repo.db.ExecContext(ctx, query, symbol, exchangeName, unit, interval, name, depth, to); err != nil {
		return err
	}
	return nil
}
