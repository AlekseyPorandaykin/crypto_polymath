package queue

import (
	"github.com/google/uuid"
	"time"
)

// Candlestick - свеча, которая пришла из очереди
type Candlestick struct {
	Exchange   string    `json:"exchange"`
	Symbol     string    `json:"symbol"`
	Unit       string    `json:"unit"`
	Interval   int       `json:"interval"`
	StartTime  time.Time `json:"start_time"`
	OpenPrice  float64   `json:"open_price"`
	HighPrice  float64   `json:"high_price"`
	LowPrice   float64   `json:"low_price"`
	ClosePrice float64   `json:"close_price"`
	Volume     float64   `json:"volume"`
}

type Indicator struct {
	Exchange string    `json:"exchange"`
	Symbol   string    `json:"symbol"`
	Unit     string    `json:"unit"`
	Interval int       `json:"interval"`
	Datetime time.Time `json:"datetime"`
	Name     string    `json:"name"`
	Depth    int       `json:"depth"`
	Value    float64   `json:"value"`
}

type CandleIndicator struct {
	Name       string    `json:"name"`
	Exchange   string    `json:"exchange"`
	Symbol     string    `json:"symbol"`
	Unit       string    `json:"unit"`
	Interval   int       `json:"interval"`
	StartTime  time.Time `json:"start_time"`
	OpenPrice  float64   `json:"open_price"`
	HighPrice  float64   `json:"high_price"`
	LowPrice   float64   `json:"low_price"`
	ClosePrice float64   `json:"close_price"`
}

type Analytic struct {
	ID             uuid.UUID `json:"id"`
	Exchange       string    `json:"exchange"`
	Symbol         string    `json:"symbol"`
	Unit           string    `json:"unit"`
	Interval       int       `json:"interval"`
	Name           string    `json:"name"`
	Datetime       time.Time `json:"datetime"`
	Depth          int       `json:"depth"`
	ByIndicator    string    `json:"by_indicator"`
	IndicatorDepth int       `json:"indicator_depth"`
	Value          float64   `json:"value"`
}

// Action - произошло какое-то действие, которое нужно обработать
type Action struct {
	Name      string        `json:"name"`
	Exchange  string        `json:"exchange"`
	Symbol    string        `json:"symbol"`
	Unit      string        `json:"unit"`
	Interval  int           `json:"interval"`
	CreatedAt time.Time     `json:"created_at"`
	Duration  time.Duration `json:"duration"`
}
