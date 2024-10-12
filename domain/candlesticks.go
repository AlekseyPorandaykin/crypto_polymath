package domain

import (
	"github.com/duke-git/lancet/v2/slice"
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
