package v1

import (
	"context"
	"fmt"
	"testing"
)

func TestClient_Server(t *testing.T) {
	c, err := NewClient("http://37.1.216.169")
	if err != nil {
		fmt.Println(err)
		return
	}
	c.Server(context.TODO())
}
func TestClient_PricesByExchange(t *testing.T) {
	c, err := NewClient("http://37.1.216.169")
	if err != nil {
		fmt.Println(err)
		return
	}
	c.PricesByExchange(context.TODO(), "bybit")
}

func TestClient_PricesBySymbol(t *testing.T) {
	c, err := NewClient("http://37.1.216.169")
	if err != nil {
		fmt.Println(err)
		return
	}
	c.PricesBySymbol(context.TODO(), "BTCUSDT")
}

func TestClient_Prices(t *testing.T) {
	c, err := NewClient("http://37.1.216.169")
	if err != nil {
		fmt.Println(err)
		return
	}
	c.Prices(context.TODO(), "bybit", "BTCUSDT")
}
func TestClient_Candlesticks(t *testing.T) {
	c, err := NewClient("http://37.1.216.169")
	if err != nil {
		fmt.Println(err)
		return
	}
	c.MinuteCandlesticks(context.TODO(), "bybit", "BTCUSDT", 1)
	c.HourCandlesticks(context.TODO(), "bybit", "BTCUSDT", 1)
	c.DayCandlesticks(context.TODO(), "bybit", "BTCUSDT")
	c.WeekCandlesticks(context.TODO(), "bybit", "BTCUSDT")
	c.MonthCandlesticks(context.TODO(), "bybit", "BTCUSDT")
}

func TestClient_Indicators(t *testing.T) {
	c, err := NewClient("http://37.1.216.169")
	if err != nil {
		fmt.Println(err)
		return
	}
	c.TrendMinuteIndicators(context.TODO(), "bybit", "BTCUSDT", 1, 10)
	c.TrendHourIndicators(context.TODO(), "bybit", "BTCUSDT", 1, 10)
	c.TrendDayIndicators(context.TODO(), "bybit", "BTCUSDT", 1, 10)
	c.TrendMonthIndicators(context.TODO(), "bybit", "BTCUSDT", 1, 10)
	c.TrendWeekIndicators(context.TODO(), "bybit", "BTCUSDT", 1, 10)

	c.MaMinuteIndicators(context.TODO(), "bybit", "BTCUSDT", 1, 10)
	c.MaHourIndicators(context.TODO(), "bybit", "BTCUSDT", 1, 10)
	c.MaDayIndicators(context.TODO(), "bybit", "BTCUSDT", 1, 10)
	c.MaMonthIndicators(context.TODO(), "bybit", "BTCUSDT", 1, 10)
	c.MaWeekIndicators(context.TODO(), "bybit", "BTCUSDT", 1, 10)

	c.EmaMinuteIndicators(context.TODO(), "bybit", "BTCUSDT", 1, 10)
	c.EmaHourIndicators(context.TODO(), "bybit", "BTCUSDT", 1, 10)
	c.EmaDayIndicators(context.TODO(), "bybit", "BTCUSDT", 1, 10)
	c.EmaMonthIndicators(context.TODO(), "bybit", "BTCUSDT", 1, 10)
	c.EmaWeekIndicators(context.TODO(), "bybit", "BTCUSDT", 1, 10)

	c.TypeCandleMinuteIndicators(context.TODO(), "bybit", "BTCUSDT", 1)
	c.TypeCandleHourIndicators(context.TODO(), "bybit", "BTCUSDT", 1)
	c.TypeCandleDayIndicators(context.TODO(), "bybit", "BTCUSDT", 1)
	c.TypeCandleMonthIndicators(context.TODO(), "bybit", "BTCUSDT", 1)
	c.TypeCandleWeekIndicators(context.TODO(), "bybit", "BTCUSDT", 1)

	c.VolatilityCandlePercentMinuteIndicators(context.TODO(), "bybit", "BTCUSDT", 1)
	c.VolatilityCandlePercentHourIndicators(context.TODO(), "bybit", "BTCUSDT", 1)
	c.VolatilityCandlePercentDayIndicators(context.TODO(), "bybit", "BTCUSDT", 1)
	c.VolatilityCandlePercentMonthIndicators(context.TODO(), "bybit", "BTCUSDT", 1)
	c.VolatilityCandlePercentWeekIndicators(context.TODO(), "bybit", "BTCUSDT", 1)
}
