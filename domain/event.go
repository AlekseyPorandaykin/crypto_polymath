package domain

import "time"

const (
	CreatedCandlestickEvent         string = "CreatedCandlestickEvent"    //domain.Candlestick
	CreatedIndicatorEvent           string = "CreatedIndicatorEvent"      //domain.Indicator
	CreatedAnalyticEvent            string = "CreatedAnalyticEvent"       //analysis.Analytic
	CreateCandleIndicatorEventEvent string = "CreateCandleIndicatorEvent" //candle_indicator.Indicator

	LoadedCandlesticksForSymbolAction string = "LoadedCandlesticksForSymbolAction" //Загрузили свечи для символа
	LoadedPricesByExchangeAction      string = "LoadedPricesByExchangeAction"      //Загрузил послежние цены биржи
)

type LoadedCandlesticksActionBody struct {
	Exchange  string
	Symbol    string
	Unit      Unit
	Interval  int
	CreatedAt time.Time
	Duration  time.Duration
}

type LoadedPricesByExchangeActionBody struct {
	Exchange  string
	Symbol    string
	CreatedAt time.Time
	Duration  time.Duration
}

type CreateIndicatorEventBody struct {
	Exchange string
	Symbol   string
	Unit     Unit
	Interval int
}
