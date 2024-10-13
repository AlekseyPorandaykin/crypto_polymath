package analysis

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/google/uuid"
	"time"
)

type Analytic struct {
	ID             uuid.UUID
	Exchange       string
	Symbol         string
	Unit           domain.Unit
	Interval       int
	Name           string
	Datetime       time.Time
	Depth          int
	ByIndicator    string
	IndicatorDepth int
	Value          float64
}

func isCorrectSequence(data []Analytic) bool {
	if len(data) == 0 {
		return false
	}
	slice.SortBy(data, func(a, b Analytic) bool {
		return a.Datetime.Before(b.Datetime)
	})
	prevVal := data[0]
	for i := range data {
		if i == 0 {
			continue
		}
		if !isPrev(data[i], prevVal) {
			return false
		}
		prevVal = data[i]
	}
	return true
}

func isPrev(current, prev Analytic) bool {
	switch current.Unit {
	case domain.MinuteUnit:
		return current.Datetime.Sub(prev.Datetime) == time.Duration(current.Interval)*time.Minute
	case domain.HourUnit:
		return current.Datetime.Sub(prev.Datetime) == time.Duration(current.Interval)*time.Hour
	case domain.DayUnit:
		return current.Datetime.Sub(prev.Datetime) == time.Duration(current.Interval)*24*time.Hour
	case domain.WeekUnit:
		return current.Datetime.Sub(prev.Datetime) == time.Duration(current.Interval)*7*24*time.Hour
	case domain.MonthUnit:
		comparableDate := time.Date(
			prev.Datetime.Year(),
			prev.Datetime.Month()+time.Month(current.Interval),
			prev.Datetime.Day(),
			prev.Datetime.Hour(),
			prev.Datetime.Minute(),
			prev.Datetime.Second(),
			prev.Datetime.Nanosecond(),
			prev.Datetime.Location(),
		)
		return (current.Datetime.Sub(comparableDate)) == 0
	}
	return false
}

type CalculatorByIndicator interface {
	Name() string
	ByIndicator() string
	SupportDepth(depth int) bool
	SupportInterval(interval int) bool
	Calculate(ctx context.Context, indicator domain.Indicator, depth int) (*Analytic, error)
}

type CalculatorByAnalytic interface {
	Name() string
	ByAnalytic() string
	SupportDepth(depth int) bool
	SupportInterval(interval int) bool
	Calculate(ctx context.Context, data Analytic, depth int) (*Analytic, error)
}
