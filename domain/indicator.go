package domain

import "time"

const (
	// Первичные идентификаторы, на основе свечей.
	TrendIndicator                   = "Trend"
	MAIndicator                      = "MA"
	EMAIndicator                     = "EMA"
	TypeCandleIndicator              = "TypeCandle"              //Направление свечи
	VolatilityCandlePercentIndicator = "VolatilityCandlePercent" //Процент изменчивости свечи
	PriceChanges                     = "PriceChanges"            //Изменение цены

	//Вторичные идентификаторы, на основе первичных идентификаторов
	TrendByMAIndicator        = "TrendByMA"
	TrendByEMAIndicator       = "TrendByEMA"
	RatioCandleToMAIndicator  = "RatioCandleToMA"
	RatioCandleToEMAIndicator = "RatioCandleToEMA"
)

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
