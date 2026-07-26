// helpers_test.go — unit-тесты внутренних функций пакета calculators.
//
// Зачем: тестируют приватные helper-функции, которые используются
// всеми аналитическими калькуляторами. Ошибка в calcTrend или
// calcExtremesIndicators каскадно ломает TrendByMA, TrendByEMA, MACD.
//
// Что покрывают:
// - calcTrend: определение тренда по экстремумам (up/down/flat)
// - calcExtremesIndicators: нахождение min/max из slice индикаторов
// - lenBatch: размер батча для расчёта экстремумов (зависит от depth)
// - compareValues: сравнение с дельтой (для stochastic)
// - Метаданные калькуляторов: Name, SupportDepth, SupportInterval
package calculators

import (
	"testing"

	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

// === calcTrend Tests ===

func TestCalcTrend_upward(t *testing.T) {
	maxValues := []float64{100, 110, 120, 130, 140}
	minValues := []float64{90, 90, 90, 90, 90}

	result := calcTrend(maxValues, minValues)
	if result != domain.UpwardTrend {
		t.Fatalf("expected UpwardTrend (1), got %d", result)
	}
}

func TestCalcTrend_downward(t *testing.T) {
	maxValues := []float64{100, 100, 100, 100, 100}
	minValues := []float64{90, 80, 70, 60, 50}

	result := calcTrend(maxValues, minValues)
	if result != domain.DownwardTrend {
		t.Fatalf("expected DownwardTrend (-1), got %d", result)
	}
}

func TestCalcTrend_flat_equal(t *testing.T) {
	maxValues := []float64{100, 110, 120}
	minValues := []float64{90, 80, 70}

	result := calcTrend(maxValues, minValues)
	if result != domain.FlatTrend {
		t.Fatalf("expected FlatTrend (0) for equal up/down, got %d", result)
	}
}

func TestCalcTrend_flat_constant(t *testing.T) {
	maxValues := []float64{100, 100, 100, 100}
	minValues := []float64{90, 90, 90, 90}

	result := calcTrend(maxValues, minValues)
	if result != domain.FlatTrend {
		t.Fatalf("expected FlatTrend for constant, got %d", result)
	}
}

func TestCalcTrend_belowThreshold(t *testing.T) {
	maxValues := []float64{100, 110, 100, 100, 100, 100, 100, 100, 100, 100}
	minValues := []float64{90, 90, 90, 90, 90, 90, 90, 90, 90, 90}

	result := calcTrend(maxValues, minValues)
	if result != domain.FlatTrend {
		t.Fatalf("expected FlatTrend below threshold, got %d", result)
	}
}

// === calcExtremesIndicators Tests ===

func TestCalcExtremesIndicators_basic(t *testing.T) {
	data := []domain.Indicator{
		{Value: 10},
		{Value: 50},
		{Value: 30},
		{Value: 5},
	}
	maxVal, minVal := calcExtremesIndicators(data)
	if maxVal != 50 {
		t.Fatalf("expected max 50, got %v", maxVal)
	}
	if minVal != 5 {
		t.Fatalf("expected min 5, got %v", minVal)
	}
}

func TestCalcExtremesIndicators_single(t *testing.T) {
	data := []domain.Indicator{{Value: 42}}
	maxVal, minVal := calcExtremesIndicators(data)
	if maxVal != 42 || minVal != 42 {
		t.Fatalf("expected 42/42, got %v/%v", maxVal, minVal)
	}
}

func TestCalcExtremesIndicators_empty(t *testing.T) {
	maxVal, minVal := calcExtremesIndicators(nil)
	if maxVal != 0 || minVal != 0 {
		t.Fatalf("expected 0/0 for empty, got %v/%v", maxVal, minVal)
	}
}

// === lenBatch Tests ===

func TestLenBatch(t *testing.T) {
	tests := []struct {
		count int
		want  int
	}{
		{5, 3},
		{10, 3},
		{15, 3},
		{16, 4},
		{20, 4},
		{21, 5},
		{49, 5},
		{50, 10},
		{100, 10},
	}
	for _, tt := range tests {
		got := lenBatch(tt.count)
		if got != tt.want {
			t.Fatalf("lenBatch(%d) = %d, want %d", tt.count, got, tt.want)
		}
	}
}

// === compareValues Tests ===

func TestCompareValues(t *testing.T) {
	if compareValues(10, 5) != 1 {
		t.Fatal("expected 1 for left > right")
	}
	if compareValues(5, 10) != -1 {
		t.Fatal("expected -1 for left < right")
	}
	if compareValues(7, 7) != 0 {
		t.Fatal("expected 0 for equal")
	}
}

// === Metadata Tests for Calculators ===

func TestTrendByEMA_Name(t *testing.T) {
	calc := NewTrendByEMA(nil)
	if calc.Name() != domain.TrendByEMAIndicator {
		t.Fatalf("expected %s, got %s", domain.TrendByEMAIndicator, calc.Name())
	}
}

func TestTrendByEMA_ByIndicator(t *testing.T) {
	calc := NewTrendByEMA(nil)
	if calc.ByIndicator() != domain.EMAIndicator {
		t.Fatalf("expected %s, got %s", domain.EMAIndicator, calc.ByIndicator())
	}
}

func TestTrendByEMA_SupportDepth(t *testing.T) {
	calc := NewTrendByEMA(nil)
	if calc.SupportDepth(9) {
		t.Fatal("depth 9 must not be supported")
	}
	if !calc.SupportDepth(10) {
		t.Fatal("depth 10 must be supported")
	}
}

func TestTrendByMA_Name(t *testing.T) {
	calc := NewTrendByMA(nil)
	if calc.Name() != domain.TrendByMAIndicator {
		t.Fatalf("expected %s, got %s", domain.TrendByMAIndicator, calc.Name())
	}
}

func TestTrendByMA_ByIndicator(t *testing.T) {
	calc := NewTrendByMA(nil)
	if calc.ByIndicator() != domain.MAIndicator {
		t.Fatalf("expected %s, got %s", domain.MAIndicator, calc.ByIndicator())
	}
}

func TestTrendByMA_SupportDepth(t *testing.T) {
	calc := NewTrendByMA(nil)
	if calc.SupportDepth(9) {
		t.Fatal("depth 9 must not be supported")
	}
	if !calc.SupportDepth(10) {
		t.Fatal("depth 10 must be supported")
	}
}

func TestRSI_Name(t *testing.T) {
	calc := NewRSI(nil)
	if calc.Name() != domain.RSIIndicator {
		t.Fatalf("expected %s, got %s", domain.RSIIndicator, calc.Name())
	}
}

func TestRSI_ByIndicator(t *testing.T) {
	calc := NewRSI(nil)
	if calc.ByIndicator() != domain.EMAIndicator {
		t.Fatalf("expected %s, got %s", domain.EMAIndicator, calc.ByIndicator())
	}
}

func TestRSI_SupportDepth(t *testing.T) {
	calc := NewRSI(nil)
	if calc.SupportDepth(9) {
		t.Fatal("depth 9 must not be supported")
	}
	if !calc.SupportDepth(10) {
		t.Fatal("depth 10 must be supported")
	}
}

func TestStochasticSignalLine_Name(t *testing.T) {
	calc := NewStochasticSignalLine(nil)
	if calc.Name() != domain.StochasticSignalLine {
		t.Fatalf("expected %s, got %s", domain.StochasticSignalLine, calc.Name())
	}
}

func TestStochasticSignalLine_ByIndicator(t *testing.T) {
	calc := NewStochasticSignalLine(nil)
	if calc.ByIndicator() != domain.StochasticMainLine {
		t.Fatalf("expected %s, got %s", domain.StochasticMainLine, calc.ByIndicator())
	}
}

func TestStochasticSignalLine_SupportDepth(t *testing.T) {
	calc := NewStochasticSignalLine(nil)
	if !calc.SupportDepth(3) {
		t.Fatal("depth 3 must be supported")
	}
	if calc.SupportDepth(5) {
		t.Fatal("depth 5 must not be supported")
	}
}

func TestRatioCandleToMA_Name(t *testing.T) {
	calc := NewRationCandleToMA(nil)
	if calc.Name() != domain.RatioCandleToMAIndicator {
		t.Fatalf("expected %s, got %s", domain.RatioCandleToMAIndicator, calc.Name())
	}
}

func TestRatioCandleToMA_SupportDepth(t *testing.T) {
	calc := NewRationCandleToMA(nil)
	if !calc.SupportDepth(1) {
		t.Fatal("depth 1 must be supported")
	}
	if calc.SupportDepth(2) {
		t.Fatal("depth 2 must not be supported")
	}
}
