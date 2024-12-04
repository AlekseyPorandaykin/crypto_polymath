package candle_indicator

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/util"
	"math"
	"time"
)

type Indicator struct {
	Name       string
	Exchange   string
	Symbol     string
	Unit       domain.Unit
	Interval   int
	StartTime  time.Time
	OpenPrice  float64
	HighPrice  float64
	LowPrice   float64
	ClosePrice float64
}

func (i Indicator) SizeBody() float64 {
	return math.Abs(i.ClosePrice - i.OpenPrice)
}

func (i Indicator) Size() float64 {
	return i.HighPrice - i.LowPrice
}

func (i Indicator) SizeBodyInPercent() float64 {
	return i.SizeBody() / i.Size() * 100
}

func (i Indicator) CloseLocation() float64 {
	return util.RoundCoin(math.Abs(i.ClosePrice-i.LowPrice)/(i.HighPrice-i.LowPrice)*100, 4)
}
func (i Indicator) OpenLocation() float64 {
	return util.RoundCoin(math.Abs(i.OpenPrice-i.LowPrice)/(i.HighPrice-i.LowPrice)*100, 4)
}

func (i Indicator) IsUp() bool {
	return i.ClosePrice > i.OpenPrice
}

func (i Indicator) IsDown() bool {
	return i.ClosePrice < i.OpenPrice
}

func (i Indicator) Direction() domain.Direction {
	if i.ClosePrice > i.OpenPrice {
		return domain.UpDirection
	}
	if i.ClosePrice < i.OpenPrice {
		return domain.DownDirection
	}
	return domain.IndefiniteDirection
}

func (i Indicator) PrevStartTime() time.Time {
	return domain.PevSequenceTime(i.Unit, i.Interval, i.StartTime)
}

type Calculator interface {
	Name() string
	Calculate(ctx context.Context, candle domain.Candlestick) (*Indicator, error)
}
