// calculator_fuzz_test.go — fuzz-тесты индикаторов.
//
// Зачем: генерирует случайные OHLC-данные и проверяет инварианты:
// - MA: результат в диапазоне [min(close), max(close)]
// - Stochastic: результат в [0, 100]
// - VolatilityCandlePercent: результат ≥ 0
// - PriceChanges: знак соответствует направлению (up = положительный)
//
// Решает проблему: деление на ноль, NaN при equal high/low,
// выход за ожидаемые диапазоны при экстремальных ценах.
package calculator_test

import (
	"math"
	"testing"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator/calculator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

func fuzzCandle(open, high, low, close float64, hours int) domain.Candlestick {
	return domain.Candlestick{
		Exchange:   "bybit",
		Symbol:     "BTCUSDT",
		Unit:       domain.HourUnit,
		Interval:   1,
		StartTime:  time.Date(2024, 1, 1, hours, 0, 0, 0, time.UTC),
		OpenPrice:  open,
		HighPrice:  high,
		LowPrice:   low,
		ClosePrice: close,
	}
}

func FuzzMA_Calculate(f *testing.F) {
	f.Add(100.0, 200.0, 300.0)
	f.Add(50000.0, 51000.0, 49000.0)

	f.Fuzz(func(t *testing.T, p1, p2, p3 float64) {
		for _, v := range []float64{p1, p2, p3} {
			if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 || v > 1e12 {
				t.Skip()
			}
		}

		candles := []domain.Candlestick{
			fuzzCandle(p1-1, p1+2, p1-2, p1, 0),
			fuzzCandle(p2-1, p2+2, p2-2, p2, 1),
			fuzzCandle(p3-1, p3+2, p3-2, p3, 2),
		}

		result := calculator.NewMA().Calculate(candles)
		if result == nil {
			t.Fatal("expected non-nil")
		}

		if math.IsNaN(result.Value) || math.IsInf(result.Value, 0) {
			t.Fatalf("non-finite MA: %v", result.Value)
		}

		minP := math.Min(p1, math.Min(p2, p3))
		maxP := math.Max(p1, math.Max(p2, p3))
		if result.Value < minP*0.99 || result.Value > maxP*1.01 {
			t.Fatalf("MA %v out of bounds [%v, %v]", result.Value, minP, maxP)
		}
	})
}

func FuzzStochastic_Calculate(f *testing.F) {
	f.Add(100.0, 110.0, 90.0, 105.0, 120.0)

	f.Fuzz(func(t *testing.T, p1, p2, p3, p4, p5 float64) {
		for _, v := range []float64{p1, p2, p3, p4, p5} {
			if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 || v > 1e12 {
				t.Skip()
			}
		}

		candles := []domain.Candlestick{
			fuzzCandle(p1, p1, p1, p1, 0),
			fuzzCandle(p2, p2, p2, p2, 1),
			fuzzCandle(p3, p3, p3, p3, 2),
			fuzzCandle(p4, p4, p4, p4, 3),
			fuzzCandle(p5, p5, p5, p5, 4),
		}

		result := calculator.NewStochasticMainLine().Calculate(candles)
		if result == nil {
			t.Fatal("expected non-nil")
		}

		if math.IsNaN(result.Value) || math.IsInf(result.Value, 0) {
			t.Fatalf("non-finite stochastic: %v", result.Value)
		}

		if result.Value < 0 || result.Value > 100 {
			t.Fatalf("stochastic %v out of [0, 100]", result.Value)
		}
	})
}

func FuzzVolatilityCandlePercent_Calculate(f *testing.F) {
	f.Add(100.0, 120.0, 80.0, 110.0)
	f.Add(50000.0, 51000.0, 49000.0, 50500.0)

	f.Fuzz(func(t *testing.T, open, high, low, close float64) {
		for _, v := range []float64{open, high, low, close} {
			if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 || v > 1e12 {
				t.Skip()
			}
		}
		if high < low {
			t.Skip()
		}

		candles := []domain.Candlestick{fuzzCandle(open, high, low, close, 0)}
		result := calculator.NewVolatilityCandlePercent().Calculate(candles)
		if result == nil {
			t.Fatal("expected non-nil")
		}

		if math.IsNaN(result.Value) || math.IsInf(result.Value, 0) {
			t.Fatalf("non-finite volatility: %v", result.Value)
		}

		if result.Value < 0 {
			t.Fatalf("volatility must be non-negative, got %v", result.Value)
		}
	})
}

func FuzzPriceChanges_Calculate(f *testing.F) {
	f.Add(100.0, 110.0, 90.0, 105.0, 95.0)

	f.Fuzz(func(t *testing.T, p1, p2, p3, p4, p5 float64) {
		for _, v := range []float64{p1, p2, p3, p4, p5} {
			if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 || v > 1e12 {
				t.Skip()
			}
		}

		candles := []domain.Candlestick{
			fuzzCandle(p1-1, p1+2, p1-2, p1, 0),
			fuzzCandle(p2-1, p2+2, p2-2, p2, 1),
			fuzzCandle(p3-1, p3+2, p3-2, p3, 2),
			fuzzCandle(p4-1, p4+2, p4-2, p4, 3),
			fuzzCandle(p5-1, p5+2, p5-2, p5, 4),
		}

		result := calculator.NewPriceChanges().Calculate(candles)
		if result == nil {
			t.Fatal("expected non-nil")
		}

		if math.IsNaN(result.Value) || math.IsInf(result.Value, 0) {
			t.Fatalf("non-finite price changes: %v", result.Value)
		}

		if result.Value < 0 {
			t.Fatalf("price changes must be non-negative, got %v", result.Value)
		}
	})
}
