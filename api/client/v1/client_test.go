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
	c.Candlesticks(context.TODO(), "bybit", "BTCUSDT", "m", 1)
	c.Candlesticks(context.TODO(), "bybit", "BTCUSDT", "H", 1)
	c.Candlesticks(context.TODO(), "bybit", "BTCUSDT", "D", 1)
	c.Candlesticks(context.TODO(), "bybit", "BTCUSDT", "W", 1)
	c.Candlesticks(context.TODO(), "bybit", "BTCUSDT", "M", 1)
}

func TestClient_Indicators(t *testing.T) {
	c, err := NewClient("http://37.1.216.169")
	if err != nil {
		fmt.Println(err)
		return
	}
	c.Indicators(context.TODO(), "bybit", "BTCUSDT", "m", 1, "Trend", 10)
	c.Indicators(context.TODO(), "bybit", "BTCUSDT", "m", 1, "MA", 10)
	c.Indicators(context.TODO(), "bybit", "BTCUSDT", "m", 1, "EMA", 10)
	c.Indicators(context.TODO(), "bybit", "BTCUSDT", "m", 1, "TypeCandle", 1)
	c.Indicators(context.TODO(), "bybit", "BTCUSDT", "m", 1, "VolatilityCandlePercent", 1)
}
