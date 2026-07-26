// golden_test.go — golden file тесты для стабильности вывода.
//
// Зачем: фиксируют точный JSON-вывод ключевых функций.
// Если формула или структура результата случайно изменится — тест упадёт,
// показывая diff между ожидаемым и фактическим выводом.
//
// Решает проблему: непреднамеренные изменения в вычислениях, которые
// не ловятся обычными assert (например, изменение порядка полей,
// потеря точности, переименование).
//
// Обновление golden-файлов: go test ./core/trading/ -run=Test.*_golden -update
package trading

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func TestSimulateAddOn_golden_longAveragingDown(t *testing.T) {
	pos := Position{Side: Long, Volume: 1, EntryPrice: 100_000, Margin: 10_000, Leverage: 10}
	add := AddOn{Price: 90_000, Margin: 10_000}
	result := Future{}.SimulateAddOn(pos, add, 0.005)

	assertGolden(t, "simulate_addon_long_avg_down", result)
}

func TestSimulateAddOn_golden_shortAveragingUp(t *testing.T) {
	pos := Position{Side: Short, Volume: 1, EntryPrice: 100_000, Margin: 10_000, Leverage: 10}
	add := AddOn{Price: 110_000, Margin: 10_000}
	result := Future{}.SimulateAddOn(pos, add, 0.005)

	assertGolden(t, "simulate_addon_short_avg_up", result)
}

func TestRiskAtPrice_golden_longInLoss(t *testing.T) {
	pos := Position{Side: Long, Volume: 1, EntryPrice: 100_000, Margin: 10_000, Leverage: 10}
	result := Future{}.RiskAtPrice(pos, 95_000, 0.005)

	assertGolden(t, "risk_at_price_long_loss", result)
}

func TestRiskAtPrice_golden_shortInProfit(t *testing.T) {
	pos := Position{Side: Short, Volume: 1, EntryPrice: 100_000, Margin: 10_000, Leverage: 10}
	result := Future{}.RiskAtPrice(pos, 95_000, 0.005)

	assertGolden(t, "risk_at_price_short_profit", result)
}

func TestDynamicStopLoss_golden(t *testing.T) {
	f := Future{Leverage: 10}
	vol := VolatilitySnapshot{RangePct: 2.5, ATRPct: 2.0, MktVol: 2.2}
	coef := DefaultStopLossCoefficients()
	result := f.DynamicStopLoss(Long, 100_000, vol, coef, VolatilityRange)

	assertGolden(t, "dynamic_stop_loss_long", result)
}

func assertGolden(t *testing.T, name string, data any) {
	t.Helper()

	golden := filepath.Join("testdata", name+".golden.json")

	got, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got = append(got, '\n')

	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden (run with -update to create): %v", err)
	}

	if string(got) != string(want) {
		t.Fatalf("output differs from golden %s:\n--- got ---\n%s\n--- want ---\n%s",
			golden, string(got), string(want))
	}
}
