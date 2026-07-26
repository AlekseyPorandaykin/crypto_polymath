// future_calculator_test.go — unit-тесты базовых расчётов фьючерсного калькулятора.
//
// Зачем: проверяют математическую корректность формул, от которых зависит
// вся торговая логика. Ошибка в средней цене или ликвидации ведёт к
// неправильному отображению позиции и потенциальным финансовым потерям.
//
// Что покрывают:
// - NewAvgEntryPrice: средневзвешенная цена входа при усреднении
// - LiquidationPrice: цена ликвидации (isolated margin) для long/short
// - UnrealizedPnL: нереализованный PnL по объёму и по залогу
// - SimulateAddOn: полная симуляция докупки (before/after snapshot)
// - VolumeFromMargin: расчёт объёма из суммы залога
// - RiskAtPrice: оценка рисков на произвольной цене
package trading

import (
	"math"
	"testing"
)

const defaultMMR = 0.005

func TestNewAvgEntryPrice(t *testing.T) {
	tests := []struct {
		name        string
		entryVolume float64
		entryPrice  float64
		newVolume   float64
		newPrice    float64
		want        float64
	}{
		{name: "equal volumes", entryVolume: 1, entryPrice: 100, newVolume: 1, newPrice: 200, want: 150},
		{name: "weighted toward larger position", entryVolume: 3, entryPrice: 100, newVolume: 1, newPrice: 200, want: 125},
		{name: "same price keeps average", entryVolume: 2, entryPrice: 50, newVolume: 4, newPrice: 50, want: 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Future{}.NewAvgEntryPrice(tt.entryVolume, tt.entryPrice, tt.newVolume, tt.newPrice)
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAvgEntryPriceBySum(t *testing.T) {
	got := Future{Leverage: 2}.NewAvgEntryPriceBySum(1, 100, 200, 200)
	want := Future{}.NewAvgEntryPrice(1, 100, 2, 200)
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if math.Abs(got-166.66666666666666) > 1e-9 {
		t.Fatalf("unexpected average: %v", got)
	}
}

func TestLiquidationPrice_long(t *testing.T) {
	// 1 BTC @ 100k, margin 10k (~10x), MMR 0.5%
	liq := Future{}.LiquidationPrice(Long, 1, 100_000, 10_000, defaultMMR)
	want := 90_452.26130653267
	if math.Abs(liq-want) > 0.01 {
		t.Fatalf("got %v, want ~%v", liq, want)
	}
}

func TestLiquidationPrice_short(t *testing.T) {
	liq := Future{}.LiquidationPrice(Short, 1, 100_000, 10_000, defaultMMR)
	want := 109_452.73631840796
	if math.Abs(liq-want) > 0.01 {
		t.Fatalf("got %v, want ~%v", liq, want)
	}
}

func TestUnrealizedPnL(t *testing.T) {
	if got := (Future{}).UnrealizedPnL(Long, 1, 0, 100, 110); got != 10 {
		t.Fatalf("long pnl: got %v", got)
	}
	if got := (Future{}).UnrealizedPnL(Short, 1, 0, 100, 110); got != -10 {
		t.Fatalf("short pnl: got %v", got)
	}
	if got := (Future{Leverage: 10}).UnrealizedPnL(Long, 0, 1000, 100, 110); math.Abs(got-1000) > 1e-9 {
		t.Fatalf("long pnl by margin: got %v", got)
	}
	if got := (Future{Leverage: 10}).UnrealizedPnL(Short, 0, 1000, 100, 110); math.Abs(got+1000) > 1e-9 {
		t.Fatalf("short pnl by margin: got %v", got)
	}
}

func TestSimulateAddOn_long_averagingDown(t *testing.T) {
	pos := Position{
		Side: Long, Volume: 1, EntryPrice: 100_000, Margin: 10_000, Leverage: 10,
	}
	add := AddOn{Price: 90_000, Margin: 10_000}

	result := Future{}.SimulateAddOn(pos, add, defaultMMR)

	if result.EntryDelta >= 0 {
		t.Fatalf("entry must decrease on long averaging down, delta=%v", result.EntryDelta)
	}
	if result.After.EntryPrice >= pos.EntryPrice {
		t.Fatalf("new entry %v must be below %v", result.After.EntryPrice, pos.EntryPrice)
	}
	if result.After.Volume <= pos.Volume {
		t.Fatalf("volume must grow: %v -> %v", pos.Volume, result.After.Volume)
	}
	if result.After.Margin != 20_000 {
		t.Fatalf("margin must be 20000, got %v", result.After.Margin)
	}
	// Ликвидация лонга при усреднении вниз смещается ниже
	if result.LiquidationAfter >= result.LiquidationBefore {
		t.Fatalf("long liq should move down: before=%v after=%v", result.LiquidationBefore, result.LiquidationAfter)
	}
	if result.BreakEvenPrice != result.After.EntryPrice {
		t.Fatalf("break even must equal new entry")
	}
}

func TestSimulateAddOn_short_averagingUp(t *testing.T) {
	pos := Position{
		Side: Short, Volume: 1, EntryPrice: 100_000, Margin: 10_000, Leverage: 10,
	}
	add := AddOn{Price: 110_000, Margin: 10_000}

	result := Future{}.SimulateAddOn(pos, add, defaultMMR)

	if result.EntryDelta <= 0 {
		t.Fatalf("entry must increase on short add when price rises, delta=%v", result.EntryDelta)
	}
	if result.LiquidationAfter <= result.LiquidationBefore {
		t.Fatalf("short liq should move up: before=%v after=%v", result.LiquidationBefore, result.LiquidationAfter)
	}
}

func TestRiskAtPrice_longInLoss(t *testing.T) {
	pos := Position{Side: Long, Volume: 1, EntryPrice: 100_000, Margin: 10_000, Leverage: 10}
	risk := Future{}.RiskAtPrice(pos, 95_000, defaultMMR)

	if risk.UnrealizedPnL != -5_000 {
		t.Fatalf("expected -5000 pnl, got %v", risk.UnrealizedPnL)
	}
	if risk.PnLPercentOnMargin != -50 {
		t.Fatalf("expected -50%% on margin, got %v", risk.PnLPercentOnMargin)
	}
	if risk.DistanceToLiquidationPct <= 0 {
		t.Fatalf("expected positive distance to liq at 95k, got %v", risk.DistanceToLiquidationPct)
	}
}

func TestVolumeFromMargin(t *testing.T) {
	got := Future{Leverage: 10}.VolumeFromMargin(1000, 50_000)
	want := 0.2
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestDistanceToLiquidationPercent(t *testing.T) {
	mark := 100_000.0
	liqLong := 90_000.0
	got := Future{}.DistanceToLiquidationPercent(Long, mark, liqLong)
	if math.Abs(got-10) > 1e-9 {
		t.Fatalf("long distance: got %v, want 10", got)
	}
	gotShort := Future{}.DistanceToLiquidationPercent(Short, mark, 110_000)
	if math.Abs(gotShort-10) > 1e-9 {
		t.Fatalf("short distance: got %v, want 10", gotShort)
	}
}
