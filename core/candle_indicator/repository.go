package candle_indicator

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/google/uuid"
	"time"
)

type StorageDTO struct {
	ID         uuid.UUID `db:"id"`
	Name       string    `db:"name"`
	Exchange   string    `db:"exchange"`
	Symbol     string    `db:"symbol"`
	Unit       string    `db:"unit"`
	Interval   int       `db:"interval"`
	StartTime  time.Time `db:"start_time"`
	OpenPrice  float64   `db:"open_price"`
	HighPrice  float64   `db:"high_price"`
	LowPrice   float64   `db:"low_price"`
	ClosePrice float64   `db:"close_price"`
	CreatedAt  time.Time `db:"created_at"`
}

type Repository interface {
	Save(ctx context.Context, data []StorageDTO) error
	Find(ctx context.Context, name, exchange, symbol, unit string, interval int, from time.Time) (*StorageDTO, error)
	FetchLast(ctx context.Context, name, exchange, symbol, unit string, interval int) ([]StorageDTO, error)
	LastAddedFromDate(ctx context.Context, name, exchange, unit string, interval int, from time.Time) ([]StorageDTO, error)
}

func DomainToStorage(data Indicator) StorageDTO {
	return StorageDTO{
		ID:         uuid.New(),
		Name:       data.Name,
		Exchange:   data.Exchange,
		Symbol:     data.Symbol,
		Unit:       string(data.Unit),
		Interval:   data.Interval,
		StartTime:  data.StartTime.In(time.UTC),
		OpenPrice:  data.OpenPrice,
		HighPrice:  data.HighPrice,
		LowPrice:   data.LowPrice,
		ClosePrice: data.ClosePrice,
		CreatedAt:  time.Now().In(time.UTC),
	}
}

func StorageToDomain(data StorageDTO) Indicator {
	return Indicator{
		Name:       data.Name,
		Exchange:   data.Exchange,
		Symbol:     data.Symbol,
		Unit:       domain.Unit(data.Unit),
		Interval:   data.Interval,
		StartTime:  data.StartTime,
		OpenPrice:  data.OpenPrice,
		HighPrice:  data.HighPrice,
		LowPrice:   data.LowPrice,
		ClosePrice: data.ClosePrice,
	}
}
