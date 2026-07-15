package calculator_test

import (
	"testing"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator/calculator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

func TestTypeCandle_Calculate_upCandle(t *testing.T) {
	dt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	calc := calculator.NewTypeCandle()

	result := calc.Calculate([]domain.Candlestick{
		{Exchange: "bybit", Symbol: "BTCUSDT", Unit: domain.HourUnit, Interval: 1, StartTime: dt, OpenPrice: 100, ClosePrice: 110},
	})
	if result == nil {
		t.Fatal("expected indicator")
	}
	if result.Value != float64(domain.UpCandle) {
		t.Fatalf("expected up candle, got %v", result.Value)
	}
	if result.Depth != 1 {
		t.Fatalf("expected depth 1, got %d", result.Depth)
	}
}

func TestTypeCandle_Calculate_downCandle(t *testing.T) {
	dt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	calc := calculator.NewTypeCandle()

	result := calc.Calculate([]domain.Candlestick{
		{Exchange: "bybit", Symbol: "BTCUSDT", Unit: domain.HourUnit, Interval: 1, StartTime: dt, OpenPrice: 110, ClosePrice: 100},
	})
	if result == nil {
		t.Fatal("expected indicator")
	}
	if result.Value != float64(domain.DownCandle) {
		t.Fatalf("expected down candle, got %v", result.Value)
	}
}

func TestTypeCandle_Calculate_emptyInput(t *testing.T) {
	calc := calculator.NewTypeCandle()
	if calc.Calculate(nil) != nil {
		t.Fatal("expected nil for empty input")
	}
}

func TestTypeCandle_SupportDepth(t *testing.T) {
	calc := calculator.NewTypeCandle()
	if !calc.SupportDepth(1) {
		t.Fatal("depth 1 must be supported")
	}
	if calc.SupportDepth(2) {
		t.Fatal("depth 2 must not be supported")
	}
}
