// stop_loss_test.go — step definitions для stop_loss.feature.
//
// Проверяет динамический стоп-лосс: адаптация к волатильности,
// trailing stop (подтягивание в прибыли), определение сработки стопа,
// расчёт цены выхода с учётом гэпов.
//
// Проблема: фиксированный SL не учитывает рыночные условия.
// Решение: SL масштабируется от волатильности + trailing для фиксации прибыли.
package bdd_test

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/cucumber/godog"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/trading"
)

// stopLossContext — состояние сценария стоп-лосса.
// Хранит параметры волатильности, результат расчёта DynamicStopLoss,
// а также промежуточные данные trailing/hit/exit.
type stopLossContext struct {
	leverage    float64
	coef        trading.StopLossCoefficients
	vol         trading.VolatilitySnapshot
	sl          trading.DynamicStopLoss
	side        trading.Side
	entry       float64
	currentStop float64
	activated   bool
	stopResult  float64
	stopHit     bool
	exitPrice   float64
}

func (sc *stopLossContext) reset() {
	*sc = stopLossContext{}
}

// Given

func (sc *stopLossContext) leverageIs(lev float64) error {
	sc.leverage = lev
	return nil
}

func (sc *stopLossContext) defaultCoefficients() error {
	sc.coef = trading.DefaultStopLossCoefficients()
	return nil
}

func (sc *stopLossContext) volatilityRangePct(pct float64) error {
	sc.vol = trading.VolatilitySnapshot{RangePct: pct}
	return nil
}

func (sc *stopLossContext) longEntryWithStop(entry, stop float64) error {
	sc.side = trading.Long
	sc.entry = entry
	sc.currentStop = stop
	return nil
}

func (sc *stopLossContext) shortEntryWithStop(entry, stop float64) error {
	sc.side = trading.Short
	sc.entry = entry
	sc.currentStop = stop
	return nil
}

func (sc *stopLossContext) trailingActivated() error {
	sc.activated = true
	return nil
}

func (sc *stopLossContext) trailingNotActivated() error {
	sc.activated = false
	return nil
}

func (sc *stopLossContext) sideStopAt(side string, stop float64) error {
	if side == "long" {
		sc.side = trading.Long
	} else {
		sc.side = trading.Short
	}
	sc.currentStop = stop
	return nil
}

// When

func (sc *stopLossContext) calcDynamicSLForLong(entry float64) error {
	sc.side = trading.Long
	sc.entry = entry
	sc.sl = trading.Future{Leverage: sc.leverage}.DynamicStopLoss(
		trading.Long, entry, sc.vol, sc.coef, trading.VolatilityRange)
	return nil
}

func (sc *stopLossContext) calcDynamicSLForShort(entry float64) error {
	sc.side = trading.Short
	sc.entry = entry
	sc.sl = trading.Future{Leverage: sc.leverage}.DynamicStopLoss(
		trading.Short, entry, sc.vol, sc.coef, trading.VolatilityRange)
	return nil
}

func (sc *stopLossContext) haCandleLowHigh(low, high float64) error {
	f := trading.Future{Leverage: sc.leverage}
	upd := f.UpdateTrailingStop(sc.side, sc.entry, sc.currentStop, low, high, sc.activated)
	sc.stopResult = upd.StopPrice
	return nil
}

func (sc *stopLossContext) candleLowHigh(low, high float64) error {
	sc.stopHit = trading.IsStopHit(sc.side, sc.currentStop, low, high)
	return nil
}

func (sc *stopLossContext) candleOpensAt(open float64) error {
	sc.exitPrice = trading.StopExitPrice(sc.side, sc.currentStop, open)
	return nil
}

// Then

func (sc *stopLossContext) slPricePctShouldBe(expected float64) error {
	if math.Abs(sc.sl.SLPricePct-expected) > 0.01 {
		return fmt.Errorf("expected SL%% %v, got %v", expected, sc.sl.SLPricePct)
	}
	return nil
}

func (sc *stopLossContext) trailPctShouldBe(expected float64) error {
	if math.Abs(sc.sl.TrailActivatePricePct-expected) > 0.01 {
		return fmt.Errorf("expected trail%% %v, got %v", expected, sc.sl.TrailActivatePricePct)
	}
	return nil
}

func (sc *stopLossContext) initStopBelow(price float64) error {
	if sc.sl.InitialStopPrice >= price {
		return fmt.Errorf("initial stop %v >= %v", sc.sl.InitialStopPrice, price)
	}
	return nil
}

func (sc *stopLossContext) initStopAbove(price float64) error {
	if sc.sl.InitialStopPrice <= price {
		return fmt.Errorf("initial stop %v <= %v", sc.sl.InitialStopPrice, price)
	}
	return nil
}

func (sc *stopLossContext) trailActivateAbove(price float64) error {
	if sc.sl.TrailActivatePrice <= price {
		return fmt.Errorf("trail activate %v <= %v", sc.sl.TrailActivatePrice, price)
	}
	return nil
}

func (sc *stopLossContext) trailActivateBelow(price float64) error {
	if sc.sl.TrailActivatePrice >= price {
		return fmt.Errorf("trail activate %v >= %v", sc.sl.TrailActivatePrice, price)
	}
	return nil
}

func (sc *stopLossContext) slMarginLoss(expected float64) error {
	if math.Abs(sc.sl.InitialSLMarginPct-expected) > 0.01 {
		return fmt.Errorf("expected SL margin %v, got %v", expected, sc.sl.InitialSLMarginPct)
	}
	return nil
}

func (sc *stopLossContext) trailMarginProfit(expected float64) error {
	if math.Abs(sc.sl.TrailActivateMarginPct-expected) > 0.01 {
		return fmt.Errorf("expected trail margin %v, got %v", expected, sc.sl.TrailActivateMarginPct)
	}
	return nil
}

func (sc *stopLossContext) stopShouldBe(expected float64) error {
	if math.Abs(sc.stopResult-expected) > 0.01 {
		return fmt.Errorf("expected stop %v, got %v", expected, sc.stopResult)
	}
	return nil
}

func (sc *stopLossContext) stopHitShouldBe(expected string) error {
	want := expected == "true"
	if sc.stopHit != want {
		return fmt.Errorf("expected stopHit=%v, got %v", want, sc.stopHit)
	}
	return nil
}

func (sc *stopLossContext) exitPriceShouldBe(expected float64) error {
	if math.Abs(sc.exitPrice-expected) > 0.01 {
		return fmt.Errorf("expected exit %v, got %v", expected, sc.exitPrice)
	}
	return nil
}

// initStopLossScenario регистрирует step definitions для stop_loss.feature:
// расчёт динамического SL, trailing stop, определение сработки, цена выхода.
func initStopLossScenario(ctx *godog.ScenarioContext) {
	sc := &stopLossContext{}

	ctx.Before(func(ctx2 context.Context, s *godog.Scenario) (context.Context, error) {
		sc.reset()
		return ctx2, nil
	})

	ctx.Step(`^leverage is ([\d.]+)x$`, sc.leverageIs)
	ctx.Step(`^default stop-loss coefficients$`, sc.defaultCoefficients)
	ctx.Step(`^volatility RangePct is ([\d.]+)%$`, sc.volatilityRangePct)
	ctx.Step(`^a long entry at ([\d.]+) with current stop at ([\d.]+)$`, sc.longEntryWithStop)
	ctx.Step(`^a short entry at ([\d.]+) with current stop at ([\d.]+)$`, sc.shortEntryWithStop)
	ctx.Step(`^trailing is activated$`, sc.trailingActivated)
	ctx.Step(`^trailing is not activated$`, sc.trailingNotActivated)
	ctx.Step(`^a (long|short) stop at ([\d.]+)$`, sc.sideStopAt)

	ctx.Step(`^I calculate dynamic stop-loss for a long at ([\d.]+)$`, sc.calcDynamicSLForLong)
	ctx.Step(`^I calculate dynamic stop-loss for a short at ([\d.]+)$`, sc.calcDynamicSLForShort)
	ctx.Step(`^HA candle has low ([\d.]+) and high ([\d.]+)$`, sc.haCandleLowHigh)
	ctx.Step(`^candle has low ([\d.]+) and high ([\d.]+)$`, sc.candleLowHigh)
	ctx.Step(`^candle opens at ([\d.]+)$`, sc.candleOpensAt)

	ctx.Step(`^SL price percent should be ([\d.]+)%$`, sc.slPricePctShouldBe)
	ctx.Step(`^trail activation percent should be ([\d.]+)%$`, sc.trailPctShouldBe)
	ctx.Step(`^initial stop price should be below ([\d.]+)$`, sc.initStopBelow)
	ctx.Step(`^initial stop price should be above ([\d.]+)$`, sc.initStopAbove)
	ctx.Step(`^trail activation price should be above ([\d.]+)$`, sc.trailActivateAbove)
	ctx.Step(`^trail activation price should be below ([\d.]+)$`, sc.trailActivateBelow)
	ctx.Step(`^SL margin loss should be (-?[\d.]+)%$`, sc.slMarginLoss)
	ctx.Step(`^trail margin profit should be ([\d.]+)%$`, sc.trailMarginProfit)
	ctx.Step(`^the stop should be ([\d.]+)$`, sc.stopShouldBe)
	ctx.Step(`^stop hit should be (true|false)$`, sc.stopHitShouldBe)
	ctx.Step(`^exit price should be ([\d.]+)$`, sc.exitPriceShouldBe)
}

// TestStopLossFeatures запускает BDD-сценарии динамического стоп-лосса.
// Покрывает: адаптивный SL, floor/cap, trailing, hit detection, exit price.
func TestStopLossFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initStopLossScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/stop_loss.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("stop-loss BDD tests failed")
	}
}
