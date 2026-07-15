package calculator

import (
	"testing"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

func BenchmarkTypeCandle_Calculate(b *testing.B) {
	calc := NewTypeCandle()
	candles := []domain.Candlestick{
		{
			Exchange: "bybit", Symbol: "BTCUSDT", Unit: domain.HourUnit, Interval: 1,
			StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			OpenPrice: 100, ClosePrice: 110, HighPrice: 115, LowPrice: 95,
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if calc.Calculate(candles) == nil {
			b.Fatal("expected indicator")
		}
	}
}
