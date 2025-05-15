package domain

import (
	"github.com/AlekseyPorandaykin/crypto_polymath/pkg/util"
	"github.com/duke-git/lancet/v2/slice"
	"math"
	"time"
)

type Unit string

const (
	MinuteUnit Unit = "m"
	HourUnit   Unit = "H"
	DayUnit    Unit = "D"
	WeekUnit   Unit = "W"
	MonthUnit  Unit = "M"
)

type Direction int

const (
	UpDirection         Direction = 1
	DownDirection       Direction = -1
	IndefiniteDirection Direction = 0
)

type Candlestick struct {
	Exchange   string
	Symbol     string
	Unit       Unit
	Interval   int
	StartTime  time.Time
	OpenPrice  float64
	HighPrice  float64
	LowPrice   float64
	ClosePrice float64
	Volume     float64
}

func (c Candlestick) SizeBody() float64 {
	return math.Abs(c.ClosePrice - c.OpenPrice)
}

func (c Candlestick) Size() float64 {
	return c.HighPrice - c.LowPrice
}

func (c Candlestick) SizeBodyInPercent() float64 {
	return util.RoundCoin(c.SizeBody()/c.Size()*100, 6)
}

func (c Candlestick) CloseLocation() float64 {
	return util.RoundCoin(math.Abs(c.ClosePrice-c.LowPrice)/(c.HighPrice-c.LowPrice)*100, 6)
}
func (c Candlestick) OpenLocation() float64 {
	return util.RoundCoin(math.Abs(c.OpenPrice-c.LowPrice)/(c.HighPrice-c.LowPrice)*100, 6)
}

func (c Candlestick) Direction() Direction {
	if c.ClosePrice > c.OpenPrice {
		return UpDirection
	}
	if c.ClosePrice < c.OpenPrice {
		return DownDirection
	}
	return IndefiniteDirection
}

func (c Candlestick) PrevStartTime() time.Time {
	return PevSequenceTime(c.Unit, c.Interval, c.StartTime)
}

func IsCorrectSequenceCandlesticks(data []Candlestick) bool {
	if len(data) == 0 {
		return false
	}
	slice.SortBy(data, func(a, b Candlestick) bool {
		return a.StartTime.Before(b.StartTime)
	})
	prevVal := data[0]
	for i := range data {
		if i == 0 {
			continue
		}
		if !isPrevCandle(data[i], prevVal) {
			return false
		}
		prevVal = data[i]
	}
	return true
}

func isPrevCandle(current, prev Candlestick) bool {
	switch current.Unit {
	case MinuteUnit:
		return current.StartTime.Sub(prev.StartTime) == time.Duration(current.Interval)*time.Minute
	case HourUnit:
		return current.StartTime.Sub(prev.StartTime) == time.Duration(current.Interval)*time.Hour
	case DayUnit:
		return current.StartTime.Sub(prev.StartTime) == time.Duration(current.Interval)*24*time.Hour
	case WeekUnit:
		return current.StartTime.Sub(prev.StartTime) == time.Duration(current.Interval)*7*24*time.Hour
	case MonthUnit:
		comparableDate := time.Date(
			prev.StartTime.Year(),
			prev.StartTime.Month()+time.Month(current.Interval),
			prev.StartTime.Day(),
			prev.StartTime.Hour(),
			prev.StartTime.Minute(),
			prev.StartTime.Second(),
			prev.StartTime.Nanosecond(),
			prev.StartTime.Location(),
		)
		return (current.StartTime.Sub(comparableDate)) == 0
	}
	return false
}

func PevSequenceTime(unit Unit, interval int, startTime time.Time) time.Time {
	switch unit {
	case MinuteUnit:
		return startTime.Add(-time.Duration(interval) * time.Minute)
	case HourUnit:
		return startTime.Add(-time.Duration(interval) * time.Hour)
	case DayUnit:
		return startTime.Add(-time.Duration(interval) * 24 * time.Hour)
	case WeekUnit:
		return startTime.Add(-time.Duration(interval) * 7 * 24 * time.Hour)
	case MonthUnit:
		return startTime.AddDate(0, -interval, 0)
	}
	return time.Time{}

}

// TODO Доделать фильтрацию
func IsOpenCandle(data Candlestick) bool {
	now := time.Now().In(time.UTC)
	startTime := data.StartTime.In(time.UTC)
	switch data.Unit {
	case MonthUnit:
		return startTime.Year() == now.Year() && startTime.Month() == now.Month()
	case WeekUnit:
		return startTime.Year() == now.Year() &&
			startTime.Month() == now.Month() &&
			startTime.Weekday() == now.Weekday()
	case DayUnit:
		return startTime.Year() == now.Year() &&
			startTime.Month() == now.Month() &&
			startTime.Day() == now.Day()
	case HourUnit:
		return startTime.Year() == now.Year() &&
			startTime.Month() == now.Month() &&
			startTime.Day() == now.Day() &&
			startTime.Hour() == now.Hour()
	case MinuteUnit:
		return startTime.Year() == now.Year() &&
			startTime.Month() == now.Month() &&
			startTime.Day() == now.Day() &&
			startTime.Hour() == now.Hour() &&
			startTime.Minute() == now.Minute()
	}
	return false
}
