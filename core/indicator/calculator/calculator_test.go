// calculator_test.go — unit-тесты первичных индикаторов (MA, EMA, Trend, Stochastic и др.)
//
// Зачем: индикаторы — фундамент аналитического конвейера.
// Ошибка в MA каскадом ломает TrendByMA, MACD и все торговые сигналы.
// Тесты проверяют:
// - Корректность вычислений на известных данных
// - Граничные случаи (пустые данные, единственная свеча, константный ряд)
// - Метаданные (Name, SupportDepth, SupportInterval)
// - Диапазон значений (Stochastic: 0–100, Trend: -1/0/+1)
package calculator_test

import (
	"testing"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator/calculator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

func newCandle(exchange, symbol string, unit domain.Unit, interval int, startTime time.Time, open, high, low, close float64) domain.Candlestick {
	return domain.Candlestick{
		Exchange:   exchange,
		Symbol:     symbol,
		Unit:       unit,
		Interval:   interval,
		StartTime:  startTime,
		OpenPrice:  open,
		HighPrice:  high,
		LowPrice:   low,
		ClosePrice: close,
	}
}

func hourlyCandles(prices []float64) []domain.Candlestick {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]domain.Candlestick, len(prices))
	for i, p := range prices {
		candles[i] = newCandle("bybit", "BTCUSDT", domain.HourUnit, 1,
			base.Add(time.Duration(i)*time.Hour),
			p-1, p+2, p-2, p)
	}
	return candles
}

// === MA Tests ===

func TestMA_Calculate_simple(t *testing.T) {
	calc := calculator.NewMA()
	candles := hourlyCandles([]float64{100, 200, 300})

	result := calc.Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Name != domain.MAIndicator {
		t.Fatalf("expected name %s, got %s", domain.MAIndicator, result.Name)
	}
	want := 200.0
	if result.Value < want-1 || result.Value > want+1 {
		t.Fatalf("expected MA ~200, got %v", result.Value)
	}
	if result.Depth != 3 {
		t.Fatalf("expected depth 3, got %d", result.Depth)
	}
}

func TestMA_Calculate_singleCandle(t *testing.T) {
	calc := calculator.NewMA()
	candles := hourlyCandles([]float64{500})

	result := calc.Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil for single candle")
	}
	if result.Value != 500 {
		t.Fatalf("expected 500, got %v", result.Value)
	}
}

func TestMA_Calculate_empty(t *testing.T) {
	calc := calculator.NewMA()
	if calc.Calculate(nil) != nil {
		t.Fatal("expected nil for empty input")
	}
}

func TestMA_SupportDepth(t *testing.T) {
	calc := calculator.NewMA()
	if calc.SupportDepth(1) {
		t.Fatal("depth 1 must not be supported")
	}
	if !calc.SupportDepth(2) {
		t.Fatal("depth 2 must be supported")
	}
	if !calc.SupportDepth(100) {
		t.Fatal("depth 100 must be supported")
	}
}

func TestMA_SupportInterval(t *testing.T) {
	calc := calculator.NewMA()
	if calc.SupportInterval(0) {
		t.Fatal("interval 0 must not be supported")
	}
	if !calc.SupportInterval(1) {
		t.Fatal("interval 1 must be supported")
	}
}

func TestMA_Calculate_usesLatestStartTime(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	candles := []domain.Candlestick{
		newCandle("bybit", "ETHUSDT", domain.HourUnit, 1, base.Add(2*time.Hour), 100, 110, 90, 105),
		newCandle("bybit", "ETHUSDT", domain.HourUnit, 1, base, 98, 108, 88, 100),
		newCandle("bybit", "ETHUSDT", domain.HourUnit, 1, base.Add(1*time.Hour), 99, 109, 89, 102),
	}
	result := calculator.NewMA().Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if !result.Datetime.Equal(base.Add(2 * time.Hour)) {
		t.Fatalf("expected latest time, got %v", result.Datetime)
	}
}

// === EMA Tests ===

func TestEMA_Calculate_simple(t *testing.T) {
	calc := calculator.NewEMA()
	candles := hourlyCandles([]float64{100, 100, 100, 100, 100})

	result := calc.Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.Name != domain.EMAIndicator {
		t.Fatalf("expected name %s, got %s", domain.EMAIndicator, result.Name)
	}
	if result.Value < 99.9 || result.Value > 100.1 {
		t.Fatalf("EMA of constant series should be ~100, got %v", result.Value)
	}
}

func TestEMA_Calculate_trending(t *testing.T) {
	calc := calculator.NewEMA()
	candles := hourlyCandles([]float64{10, 20, 30, 40, 50})

	result := calc.Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.Value <= 30 {
		t.Fatalf("EMA of uptrend should be above midpoint, got %v", result.Value)
	}
	if result.Value >= 50 {
		t.Fatalf("EMA should be below last value in uptrend, got %v", result.Value)
	}
}

func TestEMA_SupportDepth(t *testing.T) {
	calc := calculator.NewEMA()
	if calc.SupportDepth(1) {
		t.Fatal("depth 1 must not be supported")
	}
	if !calc.SupportDepth(2) {
		t.Fatal("depth 2 must be supported")
	}
}

// === Trend Tests ===

func TestTrend_Calculate_uptrend(t *testing.T) {
	calc := calculator.NewTrend()
	prices := make([]float64, 20)
	for i := range prices {
		prices[i] = 100 + float64(i)*10
	}
	candles := hourlyCandles(prices)

	result := calc.Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.Value != float64(domain.UpwardTrend) {
		t.Fatalf("expected upward trend (1), got %v", result.Value)
	}
}

func TestTrend_Calculate_downtrend(t *testing.T) {
	calc := calculator.NewTrend()
	prices := make([]float64, 20)
	for i := range prices {
		prices[i] = 300 - float64(i)*10
	}
	candles := hourlyCandles(prices)

	result := calc.Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.Value != float64(domain.DownwardTrend) {
		t.Fatalf("expected downward trend (-1), got %v", result.Value)
	}
}

func TestTrend_Calculate_flat(t *testing.T) {
	calc := calculator.NewTrend()
	prices := make([]float64, 20)
	for i := range prices {
		prices[i] = 100
	}
	candles := hourlyCandles(prices)

	result := calc.Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.Value != float64(domain.FlatTrend) {
		t.Fatalf("expected flat trend (0), got %v", result.Value)
	}
}

func TestTrend_Calculate_empty(t *testing.T) {
	calc := calculator.NewTrend()
	if calc.Calculate(nil) != nil {
		t.Fatal("expected nil for empty input")
	}
}

func TestTrend_SupportDepth(t *testing.T) {
	calc := calculator.NewTrend()
	if calc.SupportDepth(9) {
		t.Fatal("depth 9 must not be supported")
	}
	if !calc.SupportDepth(10) {
		t.Fatal("depth 10 must be supported")
	}
	if !calc.SupportDepth(50) {
		t.Fatal("depth 50 must be supported")
	}
}

// === StochasticMainLine Tests ===

func TestStochastic_Calculate_midRange(t *testing.T) {
	calc := calculator.NewStochasticMainLine()
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]domain.Candlestick, 14)
	for i := range candles {
		price := 90 + float64(i)*2
		candles[i] = newCandle("bybit", "BTCUSDT", domain.HourUnit, 1,
			base.Add(time.Duration(i)*time.Hour), price, price+5, price-5, price)
	}

	result := calc.Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.Name != domain.StochasticMainLine {
		t.Fatalf("expected name %s, got %s", domain.StochasticMainLine, result.Name)
	}
	if result.Value < 0 || result.Value > 100 {
		t.Fatalf("stochastic must be in [0, 100], got %v", result.Value)
	}
}

func TestStochastic_Calculate_allSamePrice(t *testing.T) {
	calc := calculator.NewStochasticMainLine()
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]domain.Candlestick, 5)
	for i := range candles {
		candles[i] = newCandle("bybit", "BTCUSDT", domain.HourUnit, 1,
			base.Add(time.Duration(i)*time.Hour), 100, 100, 100, 100)
	}

	result := calc.Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.Value != 0 {
		t.Fatalf("expected 0 when all prices equal, got %v", result.Value)
	}
}

func TestStochastic_Calculate_atMax(t *testing.T) {
	calc := calculator.NewStochasticMainLine()
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := []domain.Candlestick{
		newCandle("bybit", "BTCUSDT", domain.HourUnit, 1, base, 50, 50, 50, 50),
		newCandle("bybit", "BTCUSDT", domain.HourUnit, 1, base.Add(time.Hour), 100, 100, 100, 100),
	}

	result := calc.Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.Value != 100 {
		t.Fatalf("expected 100 at max, got %v", result.Value)
	}
}

func TestStochastic_Calculate_singleCandle(t *testing.T) {
	calc := calculator.NewStochasticMainLine()
	if calc.Calculate(hourlyCandles([]float64{100})) != nil {
		t.Fatal("expected nil for single candle (depth < 2)")
	}
}

func TestStochastic_SupportDepth(t *testing.T) {
	calc := calculator.NewStochasticMainLine()
	if calc.SupportDepth(1) {
		t.Fatal("depth 1 must not be supported")
	}
	if !calc.SupportDepth(14) {
		t.Fatal("depth 14 must be supported")
	}
}

// === PriceChanges Tests ===

func TestPriceChanges_Calculate_constant(t *testing.T) {
	calc := calculator.NewPriceChanges()
	candles := hourlyCandles([]float64{100, 100, 100, 100, 100})

	result := calc.Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.Value != 0 {
		t.Fatalf("expected 0 for constant prices, got %v", result.Value)
	}
}

func TestPriceChanges_Calculate_volatile(t *testing.T) {
	calc := calculator.NewPriceChanges()
	candles := hourlyCandles([]float64{100, 110, 90, 120, 80})

	result := calc.Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.Value <= 0 {
		t.Fatalf("expected positive price changes for volatile series, got %v", result.Value)
	}
	if result.Name != domain.PriceChanges {
		t.Fatalf("expected name %s, got %s", domain.PriceChanges, result.Name)
	}
}

func TestPriceChanges_Calculate_empty(t *testing.T) {
	calc := calculator.NewPriceChanges()
	if calc.Calculate(nil) != nil {
		t.Fatal("expected nil for empty input")
	}
}

func TestPriceChanges_SupportDepth(t *testing.T) {
	calc := calculator.NewPriceChanges()
	if calc.SupportDepth(1) {
		t.Fatal("depth 1 must not be supported")
	}
	if !calc.SupportDepth(5) {
		t.Fatal("depth 5 must be supported")
	}
}

// === VolatilityCandlePercent Tests ===

func TestVolatilityCandlePercent_Calculate_up(t *testing.T) {
	calc := calculator.NewVolatilityCandlePercent()
	dt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	candles := []domain.Candlestick{
		newCandle("bybit", "BTCUSDT", domain.HourUnit, 1, dt, 100, 120, 90, 110),
	}

	result := calc.Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.Value <= 0 {
		t.Fatalf("expected positive volatility for up candle, got %v", result.Value)
	}
	if result.Name != domain.VolatilityCandlePercentIndicator {
		t.Fatalf("expected name %s, got %s", domain.VolatilityCandlePercentIndicator, result.Name)
	}
}

func TestVolatilityCandlePercent_Calculate_down(t *testing.T) {
	calc := calculator.NewVolatilityCandlePercent()
	dt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	candles := []domain.Candlestick{
		newCandle("bybit", "BTCUSDT", domain.HourUnit, 1, dt, 110, 120, 90, 100),
	}

	result := calc.Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.Value <= 0 {
		t.Fatalf("expected positive volatility for down candle, got %v", result.Value)
	}
}

func TestVolatilityCandlePercent_Calculate_doji(t *testing.T) {
	calc := calculator.NewVolatilityCandlePercent()
	dt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	candles := []domain.Candlestick{
		newCandle("bybit", "BTCUSDT", domain.HourUnit, 1, dt, 100, 120, 90, 100),
	}

	result := calc.Calculate(candles)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.Value != 0 {
		t.Fatalf("expected 0 for doji, got %v", result.Value)
	}
}

func TestVolatilityCandlePercent_SupportDepth(t *testing.T) {
	calc := calculator.NewVolatilityCandlePercent()
	if !calc.SupportDepth(1) {
		t.Fatal("depth 1 must be supported")
	}
	if calc.SupportDepth(2) {
		t.Fatal("depth 2 must not be supported")
	}
}

func TestVolatilityCandlePercent_Calculate_empty(t *testing.T) {
	calc := calculator.NewVolatilityCandlePercent()
	if calc.Calculate(nil) != nil {
		t.Fatal("expected nil for empty input")
	}
}
