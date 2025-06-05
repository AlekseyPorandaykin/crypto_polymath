package candlestick

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
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

func IsPrevCandle(data StorageDTO) bool {
	now := time.Now().In(time.UTC)
	startTime := data.StartTime.In(time.UTC)
	switch data.Unit {
	case string(domain.MonthUnit):
		if !(startTime.Year() == now.Year()) {
			return false
		}
		return int(now.Month()-startTime.Month()) == data.Interval
	case string(domain.WeekUnit):
		if !(startTime.Year() == now.Year() && startTime.Month() == now.Month()) {
			return false
		}
		return int(now.Weekday()-startTime.Weekday()) == data.Interval
	case string(domain.DayUnit):
		if !(startTime.Year() == now.Year() && startTime.Month() == now.Month()) {
			return false
		}
		return (now.Day() - startTime.Day()) == data.Interval
	case string(domain.HourUnit):
		if !(startTime.Year() == now.Year() &&
			startTime.Month() == now.Month() &&
			startTime.Day() == now.Day()) {
			return false
		}
		return (now.Hour() - startTime.Hour()) == data.Interval
	case string(domain.MinuteUnit):
		if !(startTime.Year() == now.Year() &&
			startTime.Month() == now.Month() &&
			startTime.Day() == now.Day() &&
			startTime.Hour() == now.Hour()) {
			return false
		}
		return (now.Minute() - startTime.Minute()) == data.Interval
	}
	return false
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
