package indicator

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type StorageDTO struct {
	ID        uuid.UUID `db:"id"`
	Symbol    string    `db:"symbol"`
	Exchange  string    `db:"exchange"`
	Unit      string    `db:"unit"`
	Interval  int       `db:"interval"`
	Name      string    `db:"name"`
	Datetime  time.Time `db:"datetime"`
	Depth     int       `db:"depth"`
	Value     float64   `db:"value"`
	CreatedAt time.Time `db:"created_at"`
}

type UniqDTO struct {
	Symbol   string `db:"symbol"`
	Exchange string `db:"exchange"`
	Unit     string `db:"unit"`
	Interval int    `db:"interval"`
	Name     string `db:"name"`
	Depth    int    `db:"depth"`
}

type IndicatorInfo struct {
	Unit     string
	Interval int
	Name     string
	Depth    int
}

type Repository interface {
	Save(ctx context.Context, data ...StorageDTO) error
	Find(ctx context.Context, exchange, symbol, unit string, interval int, datetime time.Time, name string, depth int) (*StorageDTO, error)
	// FindMany - Пакетный поиск по конкретным datetime, чтобы не ходить в БД по одной свече за раз.
	FindMany(ctx context.Context, exchange, symbol, unit string, interval int, name string, depth int, datetimes []time.Time) ([]StorageDTO, error)
	// FindManyByName - Пакетный поиск сразу по нескольким индикаторам (name) для одной свечи.
	FindManyByName(ctx context.Context, exchange, symbol, unit string, interval int, datetime time.Time, depth int, names []string) ([]StorageDTO, error)
	List(ctx context.Context, exchange, symbol, unit string, interval int, name string, depth, limit, offset int) ([]StorageDTO, error)
	Last(
		ctx context.Context, exchange, symbol, unit string, interval int, name string, depth int,
	) (*StorageDTO, error)
	DeleteOldRows(ctx context.Context, symbol, exchangeName, unit string, interval int, name string, depth int, to time.Time) error
	ListUniq(ctx context.Context) ([]UniqDTO, error)
	LastToDate(
		ctx context.Context, exchange, symbol, unit string, interval int, name string, depth, limit int, to time.Time,
	) ([]StorageDTO, error)
	AllIndicatorInfo(ctx context.Context) (map[string][]IndicatorInfo, error)
}
