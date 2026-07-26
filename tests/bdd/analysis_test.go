// analysis_test.go — step definitions для analysis.feature.
//
// Проверяет вспомогательные алгоритмы аналитического конвейера:
// определение тренда, размер батча, математику MACD и RSI.
//
// Проблема: calcTrend/lenBatch — приватные функции, от которых зависят
// TrendByMA, TrendByEMA, MACD. Ошибка в тренде → неверный сигнал → убыточная сделка.
// BDD тестирует бизнес-поведение: «растущие экстремумы → тренд вверх».
package bdd_test

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

type analysisContext struct {
	maxValues  []float64
	minValues  []float64
	trendValue int
	count      int
	batchSize  int
	shortEMA   float64
	longEMA    float64
	macdValue  float64
	macdMain   float64
	macdSignal float64
	histogram  float64
	emaValues  []float64
	rsiValue   float64
}

func (ac *analysisContext) reset() {
	*ac = analysisContext{}
}

// --- Trend ---

func (ac *analysisContext) maxSeries(s string) error {
	ac.maxValues = parseFloatList(s)
	return nil
}

func (ac *analysisContext) minSeries(s string) error {
	ac.minValues = parseFloatList(s)
	return nil
}

func (ac *analysisContext) calculateTrend() error {
	ac.trendValue = calcTrendBDD(ac.maxValues, ac.minValues)
	return nil
}

func (ac *analysisContext) trendShouldBeUp() error {
	if ac.trendValue != domain.UpwardTrend {
		return fmt.Errorf("expected upward(1), got %v", ac.trendValue)
	}
	return nil
}

func (ac *analysisContext) trendShouldBeDown() error {
	if ac.trendValue != domain.DownwardTrend {
		return fmt.Errorf("expected downward(-1), got %v", ac.trendValue)
	}
	return nil
}

func (ac *analysisContext) trendShouldBeFlat() error {
	if ac.trendValue != domain.FlatTrend {
		return fmt.Errorf("expected flat(0), got %v", ac.trendValue)
	}
	return nil
}

// --- LenBatch ---

func (ac *analysisContext) dataCount(count int) error {
	ac.count = count
	return nil
}

func (ac *analysisContext) batchSizeShouldBe(expected int) error {
	ac.batchSize = lenBatchBDD(ac.count)
	if ac.batchSize != expected {
		return fmt.Errorf("expected batch %d, got %d", expected, ac.batchSize)
	}
	return nil
}

// --- MACD ---

func (ac *analysisContext) shortEMAValue(val float64) error {
	ac.shortEMA = val
	return nil
}

func (ac *analysisContext) longEMAValue(val float64) error {
	ac.longEMA = val
	return nil
}

func (ac *analysisContext) calcMACDMainLine() error {
	ac.macdValue = ac.shortEMA - ac.longEMA
	return nil
}

func (ac *analysisContext) macdShouldBe(expected float64) error {
	if math.Abs(ac.macdValue-expected) > 0.01 {
		return fmt.Errorf("expected MACD %v, got %v", expected, ac.macdValue)
	}
	return nil
}

func (ac *analysisContext) macdMainLineValue(val float64) error {
	ac.macdMain = val
	return nil
}

func (ac *analysisContext) macdSignalLineValue(val float64) error {
	ac.macdSignal = val
	return nil
}

func (ac *analysisContext) calcHistogram() error {
	ac.histogram = ac.macdMain - ac.macdSignal
	return nil
}

func (ac *analysisContext) histogramShouldBe(expected float64) error {
	if math.Abs(ac.histogram-expected) > 0.01 {
		return fmt.Errorf("expected histogram %v, got %v", expected, ac.histogram)
	}
	return nil
}

// --- RSI ---

func (ac *analysisContext) emaValueSeries(s string) error {
	ac.emaValues = parseFloatList(s)
	return nil
}

func (ac *analysisContext) calcRSI() error {
	var growth, fall, prev float64
	for _, v := range ac.emaValues {
		if prev == 0 {
			prev = v
			continue
		}
		if v > prev {
			growth += v - prev
		} else {
			fall += prev - v
		}
		prev = v
	}
	var rs float64
	if growth != 0 && fall != 0 {
		rs = growth / fall
	}
	if fall == 0 && growth > 0 {
		ac.rsiValue = 100
		return nil
	}
	if growth == 0 {
		ac.rsiValue = 0
		return nil
	}
	ac.rsiValue = 100 - (100 / (1 + rs))
	return nil
}

func (ac *analysisContext) rsiShouldBe(expected float64) error {
	if math.Abs(ac.rsiValue-expected) > 0.01 {
		return fmt.Errorf("expected RSI %v, got %v", expected, ac.rsiValue)
	}
	return nil
}

// --- helpers (replicate private logic for BDD) ---

const trendThreshold = 60.0

func calcTrendBDD(maxValues, minValues []float64) int {
	var countUp, countDown int
	var prevMax, prevMin float64
	for _, v := range maxValues {
		if prevMax != 0 && v > prevMax {
			countUp++
		}
		prevMax = v
	}
	for _, v := range minValues {
		if prevMin != 0 && prevMin > v {
			countDown++
		}
		prevMin = v
	}
	var pctUp, pctDown float64
	if countUp > 0 {
		pctUp = float64(countUp) / float64(len(maxValues)) * 100
	}
	if countDown > 0 {
		pctDown = float64(countDown) / float64(len(minValues)) * 100
	}
	if countUp == countDown {
		return domain.FlatTrend
	}
	if countDown == 0 && countUp != 0 && pctUp >= trendThreshold {
		return domain.UpwardTrend
	}
	if countUp == 0 && countDown != 0 && pctDown >= trendThreshold {
		return domain.DownwardTrend
	}
	if pctUp > pctDown && pctUp >= trendThreshold {
		return domain.UpwardTrend
	}
	if pctDown > pctUp && pctDown >= trendThreshold {
		return domain.DownwardTrend
	}
	return domain.FlatTrend
}

func lenBatchBDD(count int) int {
	if count <= 15 {
		return 3
	}
	if count <= 20 {
		return 4
	}
	if count < 50 {
		return 5
	}
	return 10
}

func parseFloatList(s string) []float64 {
	parts := strings.Split(s, ",")
	result := make([]float64, 0, len(parts))
	for _, p := range parts {
		v, _ := strconv.ParseFloat(strings.TrimSpace(p), 64)
		result = append(result, v)
	}
	return result
}

func initAnalysisScenario(ctx *godog.ScenarioContext) {
	ac := &analysisContext{}

	ctx.Before(func(ctx2 context.Context, s *godog.Scenario) (context.Context, error) {
		ac.reset()
		return ctx2, nil
	})

	ctx.Step(`^серия максимумов (.+)$`, ac.maxSeries)
	ctx.Step(`^серия минимумов (.+)$`, ac.minSeries)
	ctx.Step(`^определяю тренд$`, ac.calculateTrend)
	ctx.Step(`^тренд должен быть восходящим$`, ac.trendShouldBeUp)
	ctx.Step(`^тренд должен быть нисходящим$`, ac.trendShouldBeDown)
	ctx.Step(`^тренд должен быть боковым$`, ac.trendShouldBeFlat)

	ctx.Step(`^количество данных (\d+)$`, ac.dataCount)
	ctx.Step(`^размер батча должен быть (\d+)$`, ac.batchSizeShouldBe)

	ctx.Step(`^короткая EMA\(12\) = ([\d.]+)$`, ac.shortEMAValue)
	ctx.Step(`^длинная EMA\(26\) = ([\d.]+)$`, ac.longEMAValue)
	ctx.Step(`^рассчитываю MACD Main Line$`, ac.calcMACDMainLine)
	ctx.Step(`^MACD должен быть (-?[\d.]+)$`, ac.macdShouldBe)

	ctx.Step(`^MACD Main Line = (-?[\d.]+)$`, ac.macdMainLineValue)
	ctx.Step(`^MACD Signal Line = (-?[\d.]+)$`, ac.macdSignalLineValue)
	ctx.Step(`^рассчитываю MACD Histogram$`, ac.calcHistogram)
	ctx.Step(`^Histogram должен быть (-?[\d.]+)$`, ac.histogramShouldBe)

	ctx.Step(`^серия EMA значений (.+)$`, ac.emaValueSeries)
	ctx.Step(`^рассчитываю RSI$`, ac.calcRSI)
	ctx.Step(`^RSI должен быть ([\d.]+)$`, ac.rsiShouldBe)
}

// TestAnalysisFeatures запускает BDD-сценарии аналитики.
// Покрывает: calcTrend, lenBatch, MACD (Main/Signal/Histogram), RSI.
func TestAnalysisFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initAnalysisScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/analysis.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("analysis BDD tests failed")
	}
}
