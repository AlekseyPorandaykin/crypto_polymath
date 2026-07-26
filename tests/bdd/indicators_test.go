// indicators_test.go — step definitions для indicators.feature.
//
// Проверяет корректность расчёта технических индикаторов (MA, EMA, Trend,
// Stochastic) и спотового PnL.
//
// Проблема: индикаторы — фундамент аналитики и торговых стратегий.
// Ошибка в расчёте MA/EMA каскадом ломает MACD, RSI и все решения на их основе.
// Тесты гарантируют математическую корректность базовых вычислений.
package bdd_test

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator/calculator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/trading"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

// indicatorContext — состояние сценария индикаторов.
// Хранит входные свечи и результат вычисления индикатора.
type indicatorContext struct {
	candles   []domain.Candlestick
	indicator *domain.Indicator
	spotPnL   float64
	spotPct   float64
}

func (ic *indicatorContext) reset() {
	*ic = indicatorContext{}
}

func (ic *indicatorContext) candlesWithClosePrices(prices string) error {
	parts := strings.Split(prices, ",")
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	ic.candles = make([]domain.Candlestick, 0, len(parts))
	for i, p := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return err
		}
		ic.candles = append(ic.candles, domain.Candlestick{
			Exchange:   "bybit",
			Symbol:     "BTCUSDT",
			Unit:       domain.HourUnit,
			Interval:   1,
			StartTime:  base.Add(time.Duration(i) * time.Hour),
			OpenPrice:  val - 1,
			HighPrice:  val + 5,
			LowPrice:   val - 5,
			ClosePrice: val,
		})
	}
	return nil
}

func (ic *indicatorContext) noCandlestickData() error {
	ic.candles = nil
	return nil
}

func (ic *indicatorContext) risingCandles(count int, start, step float64) error {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	ic.candles = make([]domain.Candlestick, count)
	for i := range ic.candles {
		p := start + float64(i)*step
		ic.candles[i] = domain.Candlestick{
			Exchange: "bybit", Symbol: "BTCUSDT", Unit: domain.HourUnit, Interval: 1,
			StartTime: base.Add(time.Duration(i) * time.Hour),
			OpenPrice: p - 1, HighPrice: p + 5, LowPrice: p - 5, ClosePrice: p,
		}
	}
	return nil
}

func (ic *indicatorContext) fallingCandles(count int, start, step float64) error {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	ic.candles = make([]domain.Candlestick, count)
	for i := range ic.candles {
		p := start - float64(i)*step
		ic.candles[i] = domain.Candlestick{
			Exchange: "bybit", Symbol: "BTCUSDT", Unit: domain.HourUnit, Interval: 1,
			StartTime: base.Add(time.Duration(i) * time.Hour),
			OpenPrice: p - 1, HighPrice: p + 5, LowPrice: p - 5, ClosePrice: p,
		}
	}
	return nil
}

func (ic *indicatorContext) constantCandles(count int, price float64) error {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	ic.candles = make([]domain.Candlestick, count)
	for i := range ic.candles {
		ic.candles[i] = domain.Candlestick{
			Exchange: "bybit", Symbol: "BTCUSDT", Unit: domain.HourUnit, Interval: 1,
			StartTime: base.Add(time.Duration(i) * time.Hour),
			OpenPrice: price, HighPrice: price + 5, LowPrice: price - 5, ClosePrice: price,
		}
	}
	return nil
}

func (ic *indicatorContext) calcMA() error {
	ic.indicator = calculator.NewMA().Calculate(ic.candles)
	return nil
}

func (ic *indicatorContext) calcEMA() error {
	ic.indicator = calculator.NewEMA().Calculate(ic.candles)
	return nil
}

func (ic *indicatorContext) calcTrend() error {
	ic.indicator = calculator.NewTrend().Calculate(ic.candles)
	return nil
}

func (ic *indicatorContext) calcStochastic() error {
	ic.indicator = calculator.NewStochasticMainLine().Calculate(ic.candles)
	return nil
}

func (ic *indicatorContext) valueShouldBe(expected float64) error {
	if ic.indicator == nil {
		return fmt.Errorf("indicator is nil")
	}
	if math.Abs(ic.indicator.Value-expected) > 0.01 {
		return fmt.Errorf("expected %v, got %v", expected, ic.indicator.Value)
	}
	return nil
}

func (ic *indicatorContext) valueShouldBeApprox(expected float64) error {
	if ic.indicator == nil {
		return fmt.Errorf("indicator is nil")
	}
	if math.Abs(ic.indicator.Value-expected) > 1 {
		return fmt.Errorf("expected ~%v, got %v", expected, ic.indicator.Value)
	}
	return nil
}

func (ic *indicatorContext) valueShouldBeAbove(threshold float64) error {
	if ic.indicator == nil {
		return fmt.Errorf("indicator is nil")
	}
	if ic.indicator.Value <= threshold {
		return fmt.Errorf("expected > %v, got %v", threshold, ic.indicator.Value)
	}
	return nil
}

func (ic *indicatorContext) valueShouldBeBelow(threshold float64) error {
	if ic.indicator == nil {
		return fmt.Errorf("indicator is nil")
	}
	if ic.indicator.Value >= threshold {
		return fmt.Errorf("expected < %v, got %v", threshold, ic.indicator.Value)
	}
	return nil
}

func (ic *indicatorContext) nameShouldBe(name string) error {
	if ic.indicator == nil {
		return fmt.Errorf("indicator is nil")
	}
	if ic.indicator.Name != name {
		return fmt.Errorf("expected name %q, got %q", name, ic.indicator.Name)
	}
	return nil
}

func (ic *indicatorContext) noIndicatorReturned() error {
	if ic.indicator != nil {
		return fmt.Errorf("expected nil indicator, got %+v", ic.indicator)
	}
	return nil
}

// Spot

func (ic *indicatorContext) spotVolumeAtEntry(volume, entry float64) error {
	ic.candles = nil
	ic.spotPnL = 0
	ic.spotPct = 0
	ic.indicator = &domain.Indicator{Value: entry}
	ic.spotPnL = volume
	ic.spotPct = entry
	return nil
}

func (ic *indicatorContext) spotMarketAt(mark float64) error {
	volume := ic.spotPnL
	entry := ic.spotPct
	ic.spotPnL, ic.spotPct = trading.Spot{}.PnL(volume, entry, mark)
	return nil
}

func (ic *indicatorContext) spotPnlValueShouldBe(expected float64) error {
	if math.Abs(ic.spotPnL-expected) > 0.01 {
		return fmt.Errorf("expected spot PnL %v, got %v", expected, ic.spotPnL)
	}
	return nil
}

func (ic *indicatorContext) spotPnlPercentShouldBe(expected float64) error {
	if math.Abs(ic.spotPct-expected) > 0.01 {
		return fmt.Errorf("expected spot %% %v, got %v", expected, ic.spotPct)
	}
	return nil
}

// initIndicatorScenario регистрирует step definitions для indicators.feature:
// построение свечей, расчёт MA/EMA/Trend/Stochastic, проверка значений.
func initIndicatorScenario(ctx *godog.ScenarioContext) {
	ic := &indicatorContext{}

	ctx.Before(func(ctx2 context.Context, s *godog.Scenario) (context.Context, error) {
		ic.reset()
		return ctx2, nil
	})

	ctx.Step(`^candlesticks with close prices (.+)$`, ic.candlesWithClosePrices)
	ctx.Step(`^no candlestick data$`, ic.noCandlestickData)
	ctx.Step(`^(\d+) candlesticks with prices rising from ([\d.]+) by ([\d.]+)$`, ic.risingCandles)
	ctx.Step(`^(\d+) candlesticks with prices falling from ([\d.]+) by ([\d.]+)$`, ic.fallingCandles)
	ctx.Step(`^(\d+) candlesticks with constant price ([\d.]+)$`, ic.constantCandles)
	ctx.Step(`^spot volume ([\d.]+) at entry ([\d.]+)$`, ic.spotVolumeAtEntry)

	ctx.Step(`^I calculate MA$`, ic.calcMA)
	ctx.Step(`^I calculate EMA$`, ic.calcEMA)
	ctx.Step(`^I calculate Trend$`, ic.calcTrend)
	ctx.Step(`^I calculate Stochastic$`, ic.calcStochastic)
	ctx.Step(`^market is at ([\d.]+)$`, ic.spotMarketAt)

	ctx.Step(`^the indicator value should be (-?[\d.]+)$`, ic.valueShouldBe)
	ctx.Step(`^the indicator value should be approximately ([\d.]+)$`, ic.valueShouldBeApprox)
	ctx.Step(`^the indicator value should be above ([\d.]+)$`, ic.valueShouldBeAbove)
	ctx.Step(`^the indicator value should be below ([\d.]+)$`, ic.valueShouldBeBelow)
	ctx.Step(`^the indicator name should be "([^"]+)"$`, ic.nameShouldBe)
	ctx.Step(`^no indicator should be returned$`, ic.noIndicatorReturned)
	ctx.Step(`^spot PnL value should be (-?[\d.]+)$`, ic.spotPnlValueShouldBe)
	ctx.Step(`^spot PnL percent should be (-?[\d.]+)%$`, ic.spotPnlPercentShouldBe)
}

// TestIndicatorFeatures запускает BDD-сценарии технических индикаторов.
// Покрывает: MA, EMA, Trend, Stochastic, Spot PnL.
func TestIndicatorFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initIndicatorScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/indicators.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("indicator BDD tests failed")
	}
}
