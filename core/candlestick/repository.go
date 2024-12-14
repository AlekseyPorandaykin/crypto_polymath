package candlestick

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type StorageDTO struct {
	ID         uuid.UUID `db:"id"`
	Symbol     string    `db:"symbol"`
	Exchange   string    `db:"exchange"`
	Unit       string    `db:"unit"`
	Interval   int       `db:"interval"`
	StartTime  time.Time `db:"start_time"`
	OpenPrice  float64   `db:"open_price"`
	HighPrice  float64   `db:"high_price"`
	LowPrice   float64   `db:"low_price"`
	ClosePrice float64   `db:"close_price"`
	Volume     float64   `db:"volume"`
	CreatedAt  time.Time `db:"created_at"`
}

type UniqDTO struct {
	Symbol   string `db:"symbol"`
	Exchange string `db:"exchange"`
	Unit     string `db:"unit"`
	Interval int    `db:"interval"`
}

type Repository interface {
	Save(ctx context.Context, data ...StorageDTO) error
	//Last - Получаем значения с самого последнего по дате.
	Last(ctx context.Context, exchange, symbol, unit string, interval, limit, offset int) ([]StorageDTO, error)
	DeleteOldRows(ctx context.Context, exchange, symbol, unit string, interval int, to time.Time) error
	DeletePrevRows(ctx context.Context, exchange, symbol, unit string, interval int, to time.Time) error
	LastToDate(ctx context.Context, exchange, symbol, unit string, interval, limit int, to time.Time) ([]StorageDTO, error)
	//FromDate - Получаем значения с самого раннего до последнего по дате.
	FromDate(ctx context.Context, exchange, symbol, unit string, interval, limit int, to time.Time) ([]StorageDTO, error)
	ListUniq(ctx context.Context) ([]UniqDTO, error)
}
