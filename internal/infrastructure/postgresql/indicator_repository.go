package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/jmoiron/sqlx"
)

var _ indicator.Repository = (*IndicatorRepository)(nil)

type IndicatorRepository struct {
	db *sqlx.DB
}

func NewIndicatorRepository(db *sqlx.DB) *IndicatorRepository {
	return &IndicatorRepository{db: db}
}

func (repo *IndicatorRepository) Save(ctx context.Context, data ...indicator.StorageDTO) error {
	filtered := make([]indicator.StorageDTO, 0, len(data))
	for _, item := range data {
		if !math.IsNaN(item.Value) {
			filtered = append(filtered, item)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
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
VALUES (:id,:symbol,:exchange,:unit,:interval,:name,:datetime,:depth,:value,:created_at)`
	// BindNamed вызывается один раз на весь срез: sqlx сам размножит VALUES(...) на len(filtered) групп
	// со сквозной нумерацией плейсхолдеров. Вызов BindNamed по одному элементу на строку начинал
	// нумерацию ($1..$N) заново на каждой строке, из-за чего при нескольких строках в одном запросе
	// возникала ошибка "mismatched param and argument count".
	preparedQuery, args, err := repo.db.BindNamed(query, filtered)
	if err != nil {
		return err
	}
	if _, err := repo.db.ExecContext(ctx, preparedQuery, args...); err != nil {
		return err
	}
	return nil
}

func (repo *IndicatorRepository) Find(
	ctx context.Context, exchange, symbol, unit string, interval int, datetime time.Time, name string, depth int,
) (*indicator.StorageDTO, error) {
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
	err := repo.db.GetContext(ctx, &res, repo.db.Rebind(query), exchange, symbol, unit, interval, datetime, name, depth)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &res, nil
}
func (repo *IndicatorRepository) FindMany(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth int, datetimes []time.Time,
) ([]indicator.StorageDTO, error) {
	if len(datetimes) == 0 {
		return nil, nil
	}
	query, args, err := sqlx.In(`
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
WHERE exchange = ? AND symbol = ? AND unit = ? AND interval = ? AND name = ? AND depth = ? AND datetime IN (?)`,
		exchange, symbol, unit, interval, name, depth, datetimes)
	if err != nil {
		return nil, err
	}
	res := make([]indicator.StorageDTO, 0, len(datetimes))
	if err := repo.db.SelectContext(ctx, &res, repo.db.Rebind(query), args...); err != nil {
		return nil, err
	}
	return res, nil
}

func (repo *IndicatorRepository) FindManyByName(
	ctx context.Context, exchange, symbol, unit string, interval int, datetime time.Time, depth int, names []string,
) ([]indicator.StorageDTO, error) {
	if len(names) == 0 {
		return nil, nil
	}
	query, args, err := sqlx.In(`
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
WHERE exchange = ? AND symbol = ? AND unit = ? AND interval = ? AND datetime = ? AND depth = ? AND name IN (?)`,
		exchange, symbol, unit, interval, datetime, depth, names)
	if err != nil {
		return nil, err
	}
	res := make([]indicator.StorageDTO, 0, len(names))
	if err := repo.db.SelectContext(ctx, &res, repo.db.Rebind(query), args...); err != nil {
		return nil, err
	}
	return res, nil
}

func (repo *IndicatorRepository) List(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth, limit, offset int,
) ([]indicator.StorageDTO, error) {
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
	err := repo.db.SelectContext(ctx, &res, repo.db.Rebind(query), exchange, symbol, unit, interval, name, depth, limit, offset)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return res, nil
}
func (repo *IndicatorRepository) LastToDate(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth, limit int, to time.Time,
) ([]indicator.StorageDTO, error) {
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
	err := repo.db.SelectContext(ctx, &res, repo.db.Rebind(query), exchange, symbol, unit, interval, name, depth, to, limit)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return res, nil
}
func (repo *IndicatorRepository) Last(
	ctx context.Context, exchange, symbol, unit string, interval int, name string, depth int,
) (*indicator.StorageDTO, error) {
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
	err := repo.db.GetContext(ctx, &res, repo.db.Rebind(query), exchange, symbol, unit, interval, name, depth)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (repo *IndicatorRepository) ListUniq(ctx context.Context) ([]indicator.UniqDTO, error) {
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
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (repo *IndicatorRepository) DeleteOldRows(
	ctx context.Context, symbol, exchangeName, unit string, interval int, name string, depth int, to time.Time,
) error {
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
	if _, err := repo.db.ExecContext(ctx, repo.db.Rebind(query), symbol, exchangeName, unit, interval, name, depth, to); err != nil {
		return err
	}
	return nil
}

func (repo *IndicatorRepository) AllIndicatorInfo(ctx context.Context) (map[string][]indicator.IndicatorInfo, error) {
	result := make(map[string][]indicator.IndicatorInfo)
	query := `SELECT DISTINCT unit, interval, name, depth
FROM indicators`
	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var (
			unit, name      string
			interval, depth int
		)
		if err := rows.Scan(&unit, &interval, &name, &depth); err != nil {
			return nil, err
		}
		if _, has := result[name]; !has {
			result[name] = make([]indicator.IndicatorInfo, 0, 100)
		}
		result[name] = append(result[name], indicator.IndicatorInfo{
			Unit:     unit,
			Interval: interval,
			Name:     name,
			Depth:    depth,
		},
		)
	}
	return result, nil
}
