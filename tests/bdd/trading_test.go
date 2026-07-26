// Пакет bdd_test — BDD-тесты (Cucumber/Godog) для бизнес-логики проекта.
//
// Тесты написаны в формате Gherkin (.feature файлы) и проверяют поведение
// системы с точки зрения пользователя (трейдера), а не внутренней реализации.
//
// Проблема, которую решают:
// - Документируют бизнес-требования в читаемом формате
// - Позволяют валидировать поведение без знания кода
// - Обеспечивают регрессионное тестирование бизнес-правил
package bdd_test

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/cucumber/godog"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/trading"
)

// tradingContext хранит состояние между шагами одного сценария.
// Каждый сценарий получает чистый контекст через reset().
type tradingContext struct {
	mmr           float64
	leverage      float64
	pos           trading.Position
	liqPrice      float64
	liqPriceShort float64
	liqPrice2     float64
	pnl           float64
	pnlShort      float64
	avgPrice      float64
	volume        float64
	distancePct   float64
	addOnResult   trading.AddOnResult
	riskSnapshot  trading.RiskSnapshot
}

func (tc *tradingContext) reset() {
	*tc = tradingContext{mmr: 0.005}
}

// --- Background / Given ---

func (tc *tradingContext) mmrIs(pct float64) error {
	tc.mmr = pct / 100
	return nil
}

func (tc *tradingContext) leverageIs(lev float64) error {
	tc.leverage = lev
	return nil
}

func (tc *tradingContext) aLongPositionWithVolumeAndMargin(entry, volume, margin float64) error {
	tc.pos = trading.Position{Side: trading.Long, Volume: volume, EntryPrice: entry, Margin: margin, Leverage: 10}
	return nil
}

func (tc *tradingContext) aShortPositionWithVolumeAndMargin(entry, volume, margin float64) error {
	tc.pos = trading.Position{Side: trading.Short, Volume: volume, EntryPrice: entry, Margin: margin, Leverage: 10}
	return nil
}

func (tc *tradingContext) aLongPositionWithVolumeMarginLeverage(entry, volume, margin, leverage float64) error {
	tc.pos = trading.Position{Side: trading.Long, Volume: volume, EntryPrice: entry, Margin: margin, Leverage: leverage}
	return nil
}

func (tc *tradingContext) aShortPositionWithVolumeMarginLeverage(entry, volume, margin, leverage float64) error {
	tc.pos = trading.Position{Side: trading.Short, Volume: volume, EntryPrice: entry, Margin: margin, Leverage: leverage}
	return nil
}

func (tc *tradingContext) existingPositionWithVolumeAtPrice(v1, p1 float64) error {
	tc.pos = trading.Position{Volume: v1, EntryPrice: p1}
	return nil
}

// --- When ---

func (tc *tradingContext) iAddVolumeAtPrice(v2, p2 float64) error {
	tc.avgPrice = trading.Future{}.NewAvgEntryPrice(tc.pos.Volume, tc.pos.EntryPrice, v2, p2)
	return nil
}

func (tc *tradingContext) iCalcAvgEntryPriceBySum(v1, p1, sum, newPrice float64) error {
	tc.avgPrice = trading.Future{Leverage: tc.leverage}.NewAvgEntryPriceBySum(v1, p1, sum, newPrice)
	return nil
}

func (tc *tradingContext) iCalcLiquidationPrice() error {
	tc.liqPrice = trading.Future{}.LiquidationPrice(tc.pos.Side, tc.pos.Volume, tc.pos.EntryPrice, tc.pos.Margin, tc.mmr)
	return nil
}

func (tc *tradingContext) iAlsoCalcShortLiq(entry, volume, margin float64) error {
	tc.liqPriceShort = trading.Future{}.LiquidationPrice(trading.Short, volume, entry, margin, tc.mmr)
	return nil
}

func (tc *tradingContext) iCalcLiqWithMargin(margin float64) error {
	tc.liqPrice2 = trading.Future{}.LiquidationPrice(tc.pos.Side, tc.pos.Volume, tc.pos.EntryPrice, margin, tc.mmr)
	return nil
}

func (tc *tradingContext) marketPriceIs(mark float64) error {
	tc.pnl = trading.Future{}.UnrealizedPnL(tc.pos.Side, tc.pos.Volume, 0, tc.pos.EntryPrice, mark)
	tc.pnlShort = trading.Future{}.UnrealizedPnL(trading.Short, tc.pos.Volume, 0, tc.pos.EntryPrice, mark)
	return nil
}

func (tc *tradingContext) iCalcVolumeFromMargin(margin, price float64) error {
	tc.volume = trading.Future{Leverage: tc.leverage}.VolumeFromMargin(margin, price)
	return nil
}

func (tc *tradingContext) longAtMarkWithLiq(mark, liq float64) error {
	tc.distancePct = trading.Future{}.DistanceToLiquidationPercent(trading.Long, mark, liq)
	return nil
}

func (tc *tradingContext) shortAtMarkWithLiq(mark, liq float64) error {
	tc.distancePct = trading.Future{}.DistanceToLiquidationPercent(trading.Short, mark, liq)
	return nil
}

func (tc *tradingContext) iSimulateAddOn(addPrice, addMargin float64) error {
	tc.addOnResult = trading.Future{}.SimulateAddOn(tc.pos, trading.AddOn{Price: addPrice, Margin: addMargin}, tc.mmr)
	return nil
}

func (tc *tradingContext) iCalcRiskAtPrice(mark float64) error {
	tc.riskSnapshot = trading.Future{}.RiskAtPrice(tc.pos, mark, tc.mmr)
	return nil
}

// --- Then ---

func (tc *tradingContext) avgPriceShouldBe(expected float64) error {
	if math.Abs(tc.avgPrice-expected) > 0.01 {
		return fmt.Errorf("expected avg %v, got %v", expected, tc.avgPrice)
	}
	return nil
}

func (tc *tradingContext) avgPriceShouldBeApprox(expected float64) error {
	if math.Abs(tc.avgPrice-expected) > 0.01 {
		return fmt.Errorf("expected avg ~%v, got %v", expected, tc.avgPrice)
	}
	return nil
}

func (tc *tradingContext) liqShouldBeBelow(price float64) error {
	if tc.liqPrice >= price {
		return fmt.Errorf("liquidation %v >= %v", tc.liqPrice, price)
	}
	return nil
}

func (tc *tradingContext) liqShouldBeAbove(price float64) error {
	if tc.liqPrice <= price {
		return fmt.Errorf("liquidation %v <= %v", tc.liqPrice, price)
	}
	return nil
}

func (tc *tradingContext) longLiqBelowShort() error {
	if tc.liqPrice >= tc.liqPriceShort {
		return fmt.Errorf("long liq %v >= short liq %v", tc.liqPrice, tc.liqPriceShort)
	}
	return nil
}

func (tc *tradingContext) secondLiqFurther() error {
	if tc.liqPrice2 >= tc.liqPrice {
		return fmt.Errorf("second liq %v not further than first %v", tc.liqPrice2, tc.liqPrice)
	}
	return nil
}

func (tc *tradingContext) pnlShouldBe(expected float64) error {
	if math.Abs(tc.pnl-expected) > 0.01 {
		return fmt.Errorf("expected PnL %v, got %v", expected, tc.pnl)
	}
	return nil
}

func (tc *tradingContext) longPlusShortPnlZero() error {
	sum := tc.pnl + tc.pnlShort
	if math.Abs(sum) > 0.01 {
		return fmt.Errorf("long(%v) + short(%v) = %v, expected 0", tc.pnl, tc.pnlShort, sum)
	}
	return nil
}

func (tc *tradingContext) volumeShouldBe(expected float64) error {
	if math.Abs(tc.volume-expected) > 1e-9 {
		return fmt.Errorf("expected volume %v, got %v", expected, tc.volume)
	}
	return nil
}

func (tc *tradingContext) distanceShouldBe(pct float64) error {
	if math.Abs(tc.distancePct-pct) > 0.01 {
		return fmt.Errorf("expected distance %v%%, got %v%%", pct, tc.distancePct)
	}
	return nil
}

// --- Add-on Then ---

func (tc *tradingContext) newEntryBelow(price float64) error {
	if tc.addOnResult.After.EntryPrice >= price {
		return fmt.Errorf("new entry %v >= %v", tc.addOnResult.After.EntryPrice, price)
	}
	return nil
}

func (tc *tradingContext) newEntryAbove(price float64) error {
	if tc.addOnResult.After.EntryPrice <= price {
		return fmt.Errorf("new entry %v <= %v", tc.addOnResult.After.EntryPrice, price)
	}
	return nil
}

func (tc *tradingContext) newEntryApprox(price float64) error {
	if math.Abs(tc.addOnResult.After.EntryPrice-price) > 1 {
		return fmt.Errorf("new entry %v not ~%v", tc.addOnResult.After.EntryPrice, price)
	}
	return nil
}

func (tc *tradingContext) volumeShouldIncrease() error {
	if tc.addOnResult.After.Volume <= tc.addOnResult.Before.Volume {
		return fmt.Errorf("volume did not increase: %v -> %v", tc.addOnResult.Before.Volume, tc.addOnResult.After.Volume)
	}
	return nil
}

func (tc *tradingContext) totalMarginShouldBe(expected float64) error {
	if math.Abs(tc.addOnResult.After.Margin-expected) > 0.01 {
		return fmt.Errorf("expected margin %v, got %v", expected, tc.addOnResult.After.Margin)
	}
	return nil
}

func (tc *tradingContext) liqShouldMoveLower() error {
	if tc.addOnResult.LiquidationAfter >= tc.addOnResult.LiquidationBefore {
		return fmt.Errorf("liq did not move lower: %v -> %v", tc.addOnResult.LiquidationBefore, tc.addOnResult.LiquidationAfter)
	}
	return nil
}

func (tc *tradingContext) liqShouldMoveHigher() error {
	if tc.addOnResult.LiquidationAfter <= tc.addOnResult.LiquidationBefore {
		return fmt.Errorf("liq did not move higher: %v -> %v", tc.addOnResult.LiquidationBefore, tc.addOnResult.LiquidationAfter)
	}
	return nil
}

func (tc *tradingContext) breakEvenEqualsEntry() error {
	if math.Abs(tc.addOnResult.BreakEvenPrice-tc.addOnResult.After.EntryPrice) > 0.01 {
		return fmt.Errorf("break-even %v != entry %v", tc.addOnResult.BreakEvenPrice, tc.addOnResult.After.EntryPrice)
	}
	return nil
}

func (tc *tradingContext) pnlAtAddPriceApproxZero() error {
	if math.Abs(tc.addOnResult.UnrealizedPnLAtPrice) > 1 {
		return fmt.Errorf("PnL at add price %v not ~0", tc.addOnResult.UnrealizedPnLAtPrice)
	}
	return nil
}

// --- Risk Then ---

func (tc *tradingContext) riskPnlShouldBe(expected float64) error {
	if math.Abs(tc.riskSnapshot.UnrealizedPnL-expected) > 0.01 {
		return fmt.Errorf("expected risk PnL %v, got %v", expected, tc.riskSnapshot.UnrealizedPnL)
	}
	return nil
}

func (tc *tradingContext) riskPnlOnMarginShouldBe(pct float64) error {
	if math.Abs(tc.riskSnapshot.PnLPercentOnMargin-pct) > 0.01 {
		return fmt.Errorf("expected PnL on margin %v%%, got %v%%", pct, tc.riskSnapshot.PnLPercentOnMargin)
	}
	return nil
}

func (tc *tradingContext) riskDistanceShouldBePositive() error {
	if tc.riskSnapshot.DistanceToLiquidationPct <= 0 {
		return fmt.Errorf("distance to liq not positive: %v", tc.riskSnapshot.DistanceToLiquidationPct)
	}
	return nil
}

func (tc *tradingContext) riskLeverageApprox(expected float64) error {
	if math.Abs(tc.riskSnapshot.EffectiveLeverage-expected) > 0.1 {
		return fmt.Errorf("expected leverage ~%v, got %v", expected, tc.riskSnapshot.EffectiveLeverage)
	}
	return nil
}

// initTradingScenario регистрирует все step definitions для trading.feature
// и simulate_addon.feature. Godog сопоставляет regex из Step() с текстом
// шагов в .feature файлах.
func initTradingScenario(ctx *godog.ScenarioContext) {
	tc := &tradingContext{}

	ctx.Before(func(ctx2 context.Context, sc *godog.Scenario) (context.Context, error) {
		tc.reset()
		return ctx2, nil
	})

	// Background
	ctx.Step(`^MMR is ([\d.]+)%$`, tc.mmrIs)
	ctx.Step(`^leverage is ([\d.]+)x$`, tc.leverageIs)

	// Given
	ctx.Step(`^an existing position with volume ([\d.]+) at price ([\d.]+)$`, tc.existingPositionWithVolumeAtPrice)
	ctx.Step(`^a long position at ([\d.]+) with volume ([\d.]+) and margin ([\d.]+)$`, tc.aLongPositionWithVolumeAndMargin)
	ctx.Step(`^a short position at ([\d.]+) with volume ([\d.]+) and margin ([\d.]+)$`, tc.aShortPositionWithVolumeAndMargin)
	ctx.Step(`^a long position at ([\d.]+) with volume ([\d.]+) and margin ([\d.]+) and leverage ([\d.]+)x$`, tc.aLongPositionWithVolumeMarginLeverage)
	ctx.Step(`^a short position at ([\d.]+) with volume ([\d.]+) and margin ([\d.]+) and leverage ([\d.]+)x$`, tc.aShortPositionWithVolumeMarginLeverage)
	ctx.Step(`^a long position at mark price ([\d.]+) with liquidation at ([\d.]+)$`, tc.longAtMarkWithLiq)
	ctx.Step(`^a short position at mark price ([\d.]+) with liquidation at ([\d.]+)$`, tc.shortAtMarkWithLiq)

	// When
	ctx.Step(`^I add volume ([\d.]+) at price ([\d.]+)$`, tc.iAddVolumeAtPrice)
	ctx.Step(`^I calculate avg entry price by sum with volume ([\d.]+) at ([\d.]+), adding sum ([\d.]+) at price ([\d.]+)$`, tc.iCalcAvgEntryPriceBySum)
	ctx.Step(`^I calculate the liquidation price$`, tc.iCalcLiquidationPrice)
	ctx.Step(`^I also calculate the short liquidation price at ([\d.]+) with volume ([\d.]+) and margin ([\d.]+)$`, tc.iAlsoCalcShortLiq)
	ctx.Step(`^I calculate the liquidation price with margin ([\d.]+)$`, tc.iCalcLiqWithMargin)
	ctx.Step(`^market price is ([\d.]+)$`, tc.marketPriceIs)
	ctx.Step(`^I calculate volume from margin ([\d.]+) at price ([\d.]+)$`, tc.iCalcVolumeFromMargin)
	ctx.Step(`^I simulate add-on at price ([\d.]+) with margin ([\d.]+)$`, tc.iSimulateAddOn)
	ctx.Step(`^I calculate risk at price ([\d.]+)$`, tc.iCalcRiskAtPrice)

	// Then
	ctx.Step(`^the average entry price should be ([\d.]+)$`, tc.avgPriceShouldBe)
	ctx.Step(`^the average entry price should be approximately ([\d.]+)$`, tc.avgPriceShouldBeApprox)
	ctx.Step(`^the liquidation price should be below ([\d.]+)$`, tc.liqShouldBeBelow)
	ctx.Step(`^the liquidation price should be above ([\d.]+)$`, tc.liqShouldBeAbove)
	ctx.Step(`^the long liquidation should be below the short liquidation$`, tc.longLiqBelowShort)
	ctx.Step(`^the second liquidation should be further from entry$`, tc.secondLiqFurther)
	ctx.Step(`^unrealized PnL should be (-?[\d.]+)$`, tc.pnlShouldBe)
	ctx.Step(`^long PnL plus short PnL should equal zero$`, tc.longPlusShortPnlZero)
	ctx.Step(`^the volume should be ([\d.]+)$`, tc.volumeShouldBe)
	ctx.Step(`^distance to liquidation should be ([\d.]+)%$`, tc.distanceShouldBe)

	// Add-on Then
	ctx.Step(`^the new entry price should be below ([\d.]+)$`, tc.newEntryBelow)
	ctx.Step(`^the new entry price should be above ([\d.]+)$`, tc.newEntryAbove)
	ctx.Step(`^the new entry price should be approximately ([\d.]+)$`, tc.newEntryApprox)
	ctx.Step(`^the volume should increase$`, tc.volumeShouldIncrease)
	ctx.Step(`^total margin should be ([\d.]+)$`, tc.totalMarginShouldBe)
	ctx.Step(`^the liquidation price should move lower$`, tc.liqShouldMoveLower)
	ctx.Step(`^the liquidation price should move higher$`, tc.liqShouldMoveHigher)
	ctx.Step(`^break-even should equal the new entry price$`, tc.breakEvenEqualsEntry)
	ctx.Step(`^unrealized PnL at add price should be approximately 0$`, tc.pnlAtAddPriceApproxZero)

	// Risk Then
	ctx.Step(`^risk unrealized PnL should be (-?[\d.]+)$`, tc.riskPnlShouldBe)
	ctx.Step(`^risk PnL on margin should be (-?[\d.]+)%$`, tc.riskPnlOnMarginShouldBe)
	ctx.Step(`^risk distance to liquidation should be positive$`, tc.riskDistanceShouldBePositive)
	ctx.Step(`^risk effective leverage should be approximately ([\d.]+)$`, tc.riskLeverageApprox)
}

// TestTradingFeatures запускает BDD-сценарии для фьючерсного калькулятора:
// средняя цена входа, ликвидация, PnL, объём, докупка, оценка риска.
func TestTradingFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initTradingScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/trading.feature", "features/simulate_addon.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("trading BDD tests failed")
	}
}
