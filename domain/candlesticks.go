package domain

import "time"

type Unit string

const (
	MinuteUnit Unit = "m"
	HourUnit   Unit = "H"
	DayUnit    Unit = "D"
	WeekUnit   Unit = "W"
	MonthUnit  Unit = "M"
)

type Candlestick struct {
	Symbol     string
	Exchange   string
	Unit       Unit
	Interval   int
	StartTime  time.Time
	OpenPrice  float64
	HighPrice  float64
	LowPrice   float64
	ClosePrice float64
	Volume     float64
}
