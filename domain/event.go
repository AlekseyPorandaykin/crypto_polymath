package domain

import "time"

const (
	CreatedCandlestickEvent   string = "CreatedCandlestickEvent"
	CreatedIndicatorEvent     string = "CreatedIndicatorEvent"
	CreatedAnalyticEvent      string = "CreatedAnalyticEvent"
	CreateIndicatorEventEvent string = "CreateIndicatorEventEvent"

	LoadedCandlesticksForSymbolAction string = "LoadedCandlesticksForSymbolAction" //Загрузили свечи для символа
)

type ActionBody struct {
	Exchange  string
	Symbol    string
	Unit      Unit
	Interval  int
	CreatedAt time.Time
	Duration  time.Duration
}

type CreateIndicatorEventBody struct {
	Exchange string
	Symbol   string
	Unit     Unit
	Interval int
}
