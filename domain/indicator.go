package domain

import (
	"github.com/duke-git/lancet/v2/slice"
	"time"
)

const (
	// Первичные индикаторы, на основе свечей.
	TrendIndicator                   = "Trend"
	MAIndicator                      = "MA"
	EMAIndicator                     = "EMA"
	TypeCandleIndicator              = "TypeCandle"              //Направление свечи
	VolatilityCandlePercentIndicator = "VolatilityCandlePercent" //Процент изменчивости свечи в рамках одной свечи
	PriceChanges                     = "PriceChanges"            //Изменение цены, в значениях (надо умножить на 100 для процентов). Показывает как изменяется цена. За последние n-свечей (n зависит от глубины)
	StochasticMainLine               = "StochasticMainLine"      //Стохастический осциллятор (основная линия)

	//Вторичные индикаторы, на основе первичных индикаторы
	TrendByMAIndicator        = "TrendByMA"
	TrendByEMAIndicator       = "TrendByEMA"
	RatioCandleToMAIndicator  = "RatioCandleToMA"
	RatioCandleToEMAIndicator = "RatioCandleToEMA"
	RSIIndicator              = "RSI"
	MACDMainLineIndicator     = "MACDMainLine"
	MACDSignalLineIndicator   = "MACDSignalLine"
	MACDSHistogramIndicator   = "MACDSHistogram"
	StochasticSignalLine      = "StochasticSignalLine" //Стохастический осциллятор (сигнальная линия)

	//Свечные индикаторы
	HeikenAshiIndicator = "HeikenAshi"
)

// IndicatorDescriptions — пояснения к индикаторам для внешнего мира: они уходят
// в ответ GET /api/v1/server и оттуда попадают в документацию и на страницу
// инструментов. Поэтому текст здесь на английском, как и весь интерфейс, а не на
// языке комментариев в коде.
var IndicatorDescriptions = map[string]string{
	TrendIndicator:                   "Trend",
	MAIndicator:                      "Simple moving average (MA)",
	EMAIndicator:                     "Exponential moving average (EMA)",
	TypeCandleIndicator:              "Candle direction (1 - green, -1 - red)",
	VolatilityCandlePercentIndicator: "Candle volatility percent (how much the candle height changed, in percent)",
	PriceChanges:                     "Price change. Shows how the price moved over the last n steps",
	StochasticMainLine:               "Stochastic oscillator (main line)",
	TrendByMAIndicator:               "Trend based on MA",
	TrendByEMAIndicator:              "Trend based on EMA",
	RatioCandleToMAIndicator:         "Candle to MA ratio. Tells where the candle price sits relative to MA.",
	RatioCandleToEMAIndicator:        "Candle to EMA ratio. Tells where the candle price sits relative to EMA.",
	RSIIndicator:                     "Relative Strength Index. Spots the moments when the asset price has risen or fallen too far",
	MACDMainLineIndicator:            "MACD (main line)",
	MACDSignalLineIndicator:          "MACD (signal line)",
	MACDSHistogramIndicator:          "MACD (histogram)",
	StochasticSignalLine:             "Stochastic oscillator (signal line)",
}

const (
	UpCandle   = 1
	DownCandle = -1
)

const (
	UpwardTrend   = 1
	FlatTrend     = 0
	DownwardTrend = -1
)

type Indicator struct {
	Exchange string
	Symbol   string
	Unit     Unit
	Interval int
	Datetime time.Time
	Name     string
	Depth    int
	Value    float64
}

func EMA(data []float64) float64 {
	var res float64
	weight := 2 / (float64(len(data)) + 1)
	for i, item := range data {
		if i == 0 {
			res = item
			continue
		}
		res = (item * weight) + (res * (1 - weight))
	}
	return res
}

func IsCorrectSequenceIndicators(data []Indicator) bool {
	if len(data) == 0 {
		return false
	}
	slice.SortBy(data, func(a, b Indicator) bool {
		return a.Datetime.Before(b.Datetime)
	})
	prevVal := data[0]
	for i := range data {
		if i == 0 {
			continue
		}
		if !isPrevIndicator(data[i], prevVal) {
			return false
		}
		prevVal = data[i]
	}
	return true
}

func isPrevIndicator(current, prev Indicator) bool {
	switch current.Unit {
	case MinuteUnit:
		return current.Datetime.Sub(prev.Datetime) == time.Duration(current.Interval)*time.Minute
	case HourUnit:
		return current.Datetime.Sub(prev.Datetime) == time.Duration(current.Interval)*time.Hour
	case DayUnit:
		return current.Datetime.Sub(prev.Datetime) == time.Duration(current.Interval)*24*time.Hour
	case WeekUnit:
		return current.Datetime.Sub(prev.Datetime) == time.Duration(current.Interval)*7*24*time.Hour
	case MonthUnit:
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
