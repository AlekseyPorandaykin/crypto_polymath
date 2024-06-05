package candlestick

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/google/uuid"
	"strconv"
	"time"
)

func isSupportMinute(minutes int) bool {
	for _, m := range []int{1, 3, 5, 15, 30} {
		if m == minutes {
			return true
		}
	}
	return false
}

func isSupportHours(hours int) bool {
	for _, h := range []int{1, 2, 4, 6, 12} {
		if h == hours {
			return true
		}
	}
	return false
}

type ExchangeDTO struct {
	StartTime  time.Time
	OpenPrice  string
	HighPrice  string
	LowPrice   string
	ClosePrice string
	Volume     string
}

type ExchangeLoader interface {
	LastMinuteCandlesticks(ctx context.Context, symbol string, minutes int) ([]ExchangeDTO, error)
	LastHourCandlesticks(ctx context.Context, symbol string, hours int) ([]ExchangeDTO, error)
	LastDayCandlesticks(ctx context.Context, symbol string) ([]ExchangeDTO, error)
	LastWeekCandlesticks(ctx context.Context, symbol string) ([]ExchangeDTO, error)
	LastMonthCandlesticks(ctx context.Context, symbol string) ([]ExchangeDTO, error)
}

type Saver interface {
	Save(ctx context.Context, candlesticks ...domain.Candlestick) error
}

type Candlestick interface {
	AddLoader(exchange string, loader ExchangeLoader)
	LoadCandlesticks(ctx context.Context, exchange, symbol string, unit domain.Unit, interval int) ([]domain.Candlestick, error)
	DeleteOldRows(ctx context.Context, oldValueLimit int) error

	CandlesticksToDate(ctx context.Context, exchange, symbol, unit string, minutes, limit int, to time.Time) ([]domain.Candlestick, error)
	CandlesticksFromDate(ctx context.Context, exchange, symbol, unit string, minutes, limit int, to time.Time) ([]domain.Candlestick, error)
	Candlestick(ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int) ([]domain.Candlestick, error)
}

func exchangeToDomain(data ExchangeDTO, unit domain.Unit, exchange, symbol string, interval int) (domain.Candlestick, error) {
	openPrice, err := strconv.ParseFloat(data.OpenPrice, 64)
	if err != nil {
		return domain.Candlestick{}, err
	}
	highPrice, err := strconv.ParseFloat(data.HighPrice, 64)
	if err != nil {
		return domain.Candlestick{}, err
	}
	lowPrice, err := strconv.ParseFloat(data.LowPrice, 64)
	if err != nil {
		return domain.Candlestick{}, err
	}
	closedPrice, err := strconv.ParseFloat(data.ClosePrice, 64)
	if err != nil {
		return domain.Candlestick{}, err
	}
	volume, err := strconv.ParseFloat(data.Volume, 64)
	if err != nil {
		return domain.Candlestick{}, err
	}
	return domain.Candlestick{
		Symbol:     symbol,
		Exchange:   exchange,
		Unit:       unit,
		Interval:   interval,
		StartTime:  data.StartTime.In(time.UTC),
		OpenPrice:  openPrice,
		HighPrice:  highPrice,
		LowPrice:   lowPrice,
		ClosePrice: closedPrice,
		Volume:     volume,
	}, nil
}

func domainToStorage(candle domain.Candlestick) StorageDTO {
	return StorageDTO{
		ID:         uuid.New(),
		Symbol:     candle.Symbol,
		Exchange:   candle.Exchange,
		Unit:       string(candle.Unit),
		Interval:   candle.Interval,
		StartTime:  candle.StartTime.In(time.UTC),
		OpenPrice:  candle.OpenPrice,
		HighPrice:  candle.HighPrice,
		LowPrice:   candle.LowPrice,
		ClosePrice: candle.ClosePrice,
		Volume:     candle.Volume,
		CreatedAt:  time.Now().In(time.UTC),
	}
}

func exchangeToStorage(data ExchangeDTO, unit domain.Unit, exchange, symbol string, interval int) (StorageDTO, error) {
	openPrice, err := strconv.ParseFloat(data.OpenPrice, 64)
	if err != nil {
		return StorageDTO{}, err
	}
	highPrice, err := strconv.ParseFloat(data.HighPrice, 64)
	if err != nil {
		return StorageDTO{}, err
	}
	lowPrice, err := strconv.ParseFloat(data.LowPrice, 64)
	if err != nil {
		return StorageDTO{}, err
	}
	closedPrice, err := strconv.ParseFloat(data.ClosePrice, 64)
	if err != nil {
		return StorageDTO{}, err
	}
	volume, err := strconv.ParseFloat(data.Volume, 64)
	if err != nil {
		return StorageDTO{}, err
	}
	return StorageDTO{
		ID:         uuid.New(),
		Symbol:     symbol,
		Exchange:   exchange,
		Unit:       string(unit),
		Interval:   interval,
		StartTime:  data.StartTime.In(time.UTC),
		OpenPrice:  openPrice,
		HighPrice:  highPrice,
		LowPrice:   lowPrice,
		ClosePrice: closedPrice,
		Volume:     volume,
		CreatedAt:  time.Now().In(time.UTC),
	}, nil
}

func storageToDomain(data StorageDTO) domain.Candlestick {
	return domain.Candlestick{
		Symbol:     data.Symbol,
		Exchange:   data.Exchange,
		Unit:       domain.Unit(data.Unit),
		Interval:   data.Interval,
		StartTime:  data.StartTime,
		OpenPrice:  data.OpenPrice,
		HighPrice:  data.HighPrice,
		LowPrice:   data.LowPrice,
		ClosePrice: data.ClosePrice,
		Volume:     data.Volume,
	}
}
