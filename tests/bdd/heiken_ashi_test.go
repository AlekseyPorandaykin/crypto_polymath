// heiken_ashi_test.go — step definitions для heiken_ashi.feature.
//
// Проверяет расчёт Heiken Ashi свечей и свойства типа candle_indicator.Indicator.
// HA-свечи — основа стратегии trailing stop и определения тренда.
//
// Проблема: неверный HA → неверный trailing → преждевременный выход из позиции.
package bdd_test

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/cucumber/godog"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/trading"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

type heikenAshiContext struct {
	candle   domain.Candlestick
	prevHA   *candle_indicator.Indicator
	haResult *candle_indicator.Indicator
	ind      candle_indicator.Indicator
	isDoji   bool
}

func (hc *heikenAshiContext) reset() {
	*hc = heikenAshiContext{}
}

func (hc *heikenAshiContext) candleOHLC(o, h, l, c float64) error {
	hc.candle = domain.Candlestick{
		Exchange: "bybit", Symbol: "BTCUSDT", Unit: domain.HourUnit, Interval: 1,
		OpenPrice: o, HighPrice: h, LowPrice: l, ClosePrice: c,
	}
	return nil
}

func (hc *heikenAshiContext) noPreviousHA() error {
	hc.prevHA = nil
	return nil
}

func (hc *heikenAshiContext) previousHAOpenClose(open, close float64) error {
	hc.prevHA = &candle_indicator.Indicator{OpenPrice: open, ClosePrice: close}
	return nil
}

func (hc *heikenAshiContext) calculateHA() error {
	openHA := hc.candle.OpenPrice
	if hc.prevHA != nil {
		openHA = (hc.prevHA.ClosePrice + hc.prevHA.OpenPrice) / 2
	}
	closeHA := (hc.candle.OpenPrice + hc.candle.ClosePrice + hc.candle.HighPrice + hc.candle.LowPrice) / 4
	highHA := math.Max(hc.candle.HighPrice, math.Max(openHA, closeHA))
	lowHA := math.Min(hc.candle.LowPrice, math.Min(openHA, closeHA))

	hc.haResult = &candle_indicator.Indicator{
		Name:       domain.HeikenAshiIndicator,
		Exchange:   hc.candle.Exchange,
		Symbol:     hc.candle.Symbol,
		Unit:       hc.candle.Unit,
		Interval:   hc.candle.Interval,
		OpenPrice:  openHA,
		HighPrice:  highHA,
		LowPrice:   lowHA,
		ClosePrice: closeHA,
	}
	return nil
}

func (hc *heikenAshiContext) haCloseShouldBe(expected float64) error {
	if math.Abs(hc.haResult.ClosePrice-expected) > 0.01 {
		return fmt.Errorf("expected HA Close %v, got %v", expected, hc.haResult.ClosePrice)
	}
	return nil
}

func (hc *heikenAshiContext) haOpenShouldBe(expected float64) error {
	if math.Abs(hc.haResult.OpenPrice-expected) > 0.01 {
		return fmt.Errorf("expected HA Open %v, got %v", expected, hc.haResult.OpenPrice)
	}
	return nil
}

func (hc *heikenAshiContext) haHighShouldBe(expected float64) error {
	if math.Abs(hc.haResult.HighPrice-expected) > 0.01 {
		return fmt.Errorf("expected HA High %v, got %v", expected, hc.haResult.HighPrice)
	}
	return nil
}

func (hc *heikenAshiContext) haLowShouldBe(expected float64) error {
	if math.Abs(hc.haResult.LowPrice-expected) > 0.01 {
		return fmt.Errorf("expected HA Low %v, got %v", expected, hc.haResult.LowPrice)
	}
	return nil
}

// Indicator properties

func (hc *heikenAshiContext) indicatorOHLC(o, h, l, c float64) error {
	hc.ind = candle_indicator.Indicator{OpenPrice: o, HighPrice: h, LowPrice: l, ClosePrice: c}
	return nil
}

func (hc *heikenAshiContext) bodyShouldBe(expected float64) error {
	if math.Abs(hc.ind.SizeBody()-expected) > 0.01 {
		return fmt.Errorf("expected body %v, got %v", expected, hc.ind.SizeBody())
	}
	return nil
}

func (hc *heikenAshiContext) sizeShouldBe(expected float64) error {
	if math.Abs(hc.ind.Size()-expected) > 0.01 {
		return fmt.Errorf("expected size %v, got %v", expected, hc.ind.Size())
	}
	return nil
}

func (hc *heikenAshiContext) bodyPercentShouldBe(expected float64) error {
	if math.Abs(hc.ind.SizeBodyInPercent()-expected) > 0.01 {
		return fmt.Errorf("expected body%% %v, got %v", expected, hc.ind.SizeBodyInPercent())
	}
	return nil
}

func (hc *heikenAshiContext) directionShouldBeUp() error {
	if hc.ind.Direction() != domain.UpDirection {
		return fmt.Errorf("expected Up, got %v", hc.ind.Direction())
	}
	return nil
}

func (hc *heikenAshiContext) directionShouldBeDown() error {
	if hc.ind.Direction() != domain.DownDirection {
		return fmt.Errorf("expected Down, got %v", hc.ind.Direction())
	}
	return nil
}

func (hc *heikenAshiContext) directionShouldBeIndefinite() error {
	if hc.ind.Direction() != domain.IndefiniteDirection {
		return fmt.Errorf("expected Indefinite, got %v", hc.ind.Direction())
	}
	return nil
}

// Doji

func (hc *heikenAshiContext) haCandleForDoji(o, c, h, l float64, threshold float64) error {
	hc.isDoji = trading.IsHeikenAshiDoji(o, c, h, l, threshold)
	return nil
}

func (hc *heikenAshiContext) shouldBeDoji() error {
	if !hc.isDoji {
		return fmt.Errorf("expected doji=true, got false")
	}
	return nil
}

func (hc *heikenAshiContext) shouldNotBeDoji() error {
	if hc.isDoji {
		return fmt.Errorf("expected doji=false, got true")
	}
	return nil
}

func initHeikenAshiScenario(ctx *godog.ScenarioContext) {
	hc := &heikenAshiContext{}

	ctx.Before(func(ctx2 context.Context, s *godog.Scenario) (context.Context, error) {
		hc.reset()
		return ctx2, nil
	})

	ctx.Step(`^обычная свеча с O=([\d.]+) H=([\d.]+) L=([\d.]+) C=([\d.]+)$`, hc.candleOHLC)
	ctx.Step(`^предыдущей HA свечи нет$`, hc.noPreviousHA)
	ctx.Step(`^предыдущая HA свеча имела Open=([\d.]+) Close=([\d.]+)$`, hc.previousHAOpenClose)
	ctx.Step(`^рассчитываю Heiken Ashi$`, hc.calculateHA)
	ctx.Step(`^HA Close должен быть ([\d.]+)$`, hc.haCloseShouldBe)
	ctx.Step(`^HA Open должен быть ([\d.]+)$`, hc.haOpenShouldBe)
	ctx.Step(`^HA High должен быть ([\d.]+)$`, hc.haHighShouldBe)
	ctx.Step(`^HA Low должен быть ([\d.]+)$`, hc.haLowShouldBe)

	ctx.Step(`^HA индикатор с O=([\d.]+) H=([\d.]+) L=([\d.]+) C=([\d.]+)$`, hc.indicatorOHLC)
	ctx.Step(`^размер тела должен быть ([\d.]+)$`, hc.bodyShouldBe)
	ctx.Step(`^полный размер должен быть ([\d.]+)$`, hc.sizeShouldBe)
	ctx.Step(`^тело в процентах от размера должно быть ([\d.]+)%$`, hc.bodyPercentShouldBe)
	ctx.Step(`^направление должно быть вверх$`, hc.directionShouldBeUp)
	ctx.Step(`^направление должно быть вниз$`, hc.directionShouldBeDown)
	ctx.Step(`^направление должно быть неопределённым$`, hc.directionShouldBeIndefinite)

	ctx.Step(`^HA свеча с O=([\d.]+) C=([\d.]+) H=([\d.]+) L=([\d.]+) и порог доджи ([\d.]+)%$`, hc.haCandleForDoji)
	ctx.Step(`^это доджи$`, hc.shouldBeDoji)
	ctx.Step(`^это не доджи$`, hc.shouldNotBeDoji)
}

// TestHeikenAshiFeatures запускает BDD-сценарии Heiken Ashi свечей.
// Покрывает: формулы HA OHLC, свойства Indicator, определение доджи.
func TestHeikenAshiFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initHeikenAshiScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/heiken_ashi.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("heiken ashi BDD tests failed")
	}
}
