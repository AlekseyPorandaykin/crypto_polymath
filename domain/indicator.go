package domain

import (
	"time"
)

const (
	// Первичные идентификаторы, на основе свечей.
	TrendIndicator                   = "Trend"
	MAIndicator                      = "MA"
	EMAIndicator                     = "EMA"
	TypeCandleIndicator              = "TypeCandle"              //Направление свечи
	VolatilityCandlePercentIndicator = "VolatilityCandlePercent" //Процент изменчивости свечи
	PriceChanges                     = "PriceChanges"            //Изменение цены
	StochasticMainLine               = "StochasticMainLine"      //Стохастический осциллятор (основная линия)

	//Вторичные идентификаторы, на основе первичных идентификаторов
	TrendByMAIndicator        = "TrendByMA"
	TrendByEMAIndicator       = "TrendByEMA"
	RatioCandleToMAIndicator  = "RatioCandleToMA"
	RatioCandleToEMAIndicator = "RatioCandleToEMA"
	RSIIndicator              = "RSI"
	MACDMainLineIndicator     = "MACDMainLine"
	MACDSignalLineIndicator   = "MACDSignalLine"
	MACDSHistogramIndicator   = "MACDSHistogram"
	StochasticSignalLine      = "StochasticSignalLine" //Стохастический осциллятор (сигнальная линия)
)

var IndicatorDescriptions = map[string]string{
	TrendIndicator:                   "Тренд",
	MAIndicator:                      "Простая скользящей средней (MA)",
	EMAIndicator:                     "Экспоненциальной скользящей средней (EMA)",
	TypeCandleIndicator:              "Направление свечи(1 - зеленая, -1 - красная)",
	VolatilityCandlePercentIndicator: "Процент изменчивости свечи(определяет высоту свечи)",
	PriceChanges:                     "Изменение цены. Определяем как изменялась цена последние n-шагов",
	StochasticMainLine:               "Стохастический осциллятор (основная линия)",
	TrendByMAIndicator:               "Тренд на основе MA",
	TrendByEMAIndicator:              "Тренд на основе EMA",
	RatioCandleToMAIndicator:         "Отношение свечи к MA. Для определения положения цены свечи к MA.",
	RatioCandleToEMAIndicator:        "Отношение свечи к EMA. Для определения положения цены свечи к EMA.",
	RSIIndicator:                     "Relative Strength Index. Можно определить моменты, когда цена актива выросла или упала слишком сильно",
	MACDMainLineIndicator:            "MACD (основная линия)",
	MACDSignalLineIndicator:          "MACD (сигнальная линия)",
	MACDSHistogramIndicator:          "MACD (гистограмма)",
	StochasticSignalLine:             "Стохастический осциллятор (сигнальная линия)",
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
