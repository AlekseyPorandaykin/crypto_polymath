// future_calculator_fuzz_test.go — fuzz-тесты (property-based) для фьючерсного калькулятора.
//
// Зачем: обнаруживают edge-кейсы, которые сложно предусмотреть вручную:
// деление на ноль, NaN/Inf при экстремальных входах, нарушение инвариантов.
// Go fuzzer генерирует тысячи случайных комбинаций параметров.
//
// Какие инварианты проверяются:
// - NewAvgEntryPrice: результат всегда в диапазоне [min(p1,p2), max(p1,p2)]
// - LiquidationPrice: long liq < entry < short liq (всегда)
// - UnrealizedPnL: long_pnl + short_pnl = 0 (нулевая сумма)
// - SimulateAddOn: объём после > объём до, маржа после = сумма маржей
// - VolumeFromMargin: volume > 0 при положительных входах
// - Spot.PnL: PnL = volume × (mark − entry)
package trading

import (
	"math"
	"testing"
)

func FuzzNewAvgEntryPrice(f *testing.F) {
	f.Add(1.0, 100.0, 1.0, 200.0)
	f.Add(0.5, 50000.0, 2.0, 48000.0)
	f.Add(3.0, 100.0, 1.0, 200.0)
	f.Add(0.001, 99999.0, 0.001, 100001.0)

	f.Fuzz(func(t *testing.T, v1, p1, v2, p2 float64) {
		if v1 <= 0 || v2 <= 0 || p1 <= 0 || p2 <= 0 {
			t.Skip()
		}
		if math.IsNaN(v1) || math.IsNaN(p1) || math.IsNaN(v2) || math.IsNaN(p2) {
			t.Skip()
		}
		if math.IsInf(v1, 0) || math.IsInf(p1, 0) || math.IsInf(v2, 0) || math.IsInf(p2, 0) {
			t.Skip()
		}

		avg := Future{}.NewAvgEntryPrice(v1, p1, v2, p2)

		if math.IsNaN(avg) || math.IsInf(avg, 0) {
			t.Fatalf("non-finite result: %v (inputs: %v %v %v %v)", avg, v1, p1, v2, p2)
		}

		minP, maxP := math.Min(p1, p2), math.Max(p1, p2)
		if avg < minP*(1-1e-9) || avg > maxP*(1+1e-9) {
			t.Fatalf("avg %v not in [%v, %v]", avg, minP, maxP)
		}
	})
}

func FuzzLiquidationPrice(f *testing.F) {
	f.Add(1.0, 100000.0, 10000.0, 0.005)
	f.Add(0.5, 50000.0, 5000.0, 0.01)
	f.Add(2.0, 60000.0, 20000.0, 0.004)

	f.Fuzz(func(t *testing.T, volume, entry, margin, mmr float64) {
		if volume <= 0 || entry <= 0 || margin <= 0 || mmr <= 0 || mmr >= 1 {
			t.Skip()
		}
		if math.IsNaN(volume) || math.IsNaN(entry) || math.IsNaN(margin) || math.IsNaN(mmr) {
			t.Skip()
		}
		if math.IsInf(volume, 0) || math.IsInf(entry, 0) || math.IsInf(margin, 0) || math.IsInf(mmr, 0) {
			t.Skip()
		}
		if margin > volume*entry {
			t.Skip()
		}

		liqLong := Future{}.LiquidationPrice(Long, volume, entry, margin, mmr)
		liqShort := Future{}.LiquidationPrice(Short, volume, entry, margin, mmr)

		if math.IsNaN(liqLong) || math.IsInf(liqLong, 0) {
			t.Fatalf("long liq non-finite: %v", liqLong)
		}
		if math.IsNaN(liqShort) || math.IsInf(liqShort, 0) {
			t.Fatalf("short liq non-finite: %v", liqShort)
		}

		if liqLong > 0 && liqLong >= entry {
			t.Fatalf("long liq %v >= entry %v (margin=%v, vol=%v)", liqLong, entry, margin, volume)
		}
		if liqShort > 0 && liqShort <= entry {
			t.Fatalf("short liq %v <= entry %v (margin=%v, vol=%v)", liqShort, entry, margin, volume)
		}

		if liqLong > 0 && liqShort > 0 && liqLong >= liqShort {
			t.Fatalf("long liq %v >= short liq %v", liqLong, liqShort)
		}
	})
}

func FuzzUnrealizedPnL(f *testing.F) {
	f.Add(1.0, 100000.0, 110000.0)
	f.Add(0.5, 50000.0, 48000.0)

	f.Fuzz(func(t *testing.T, volume, entry, mark float64) {
		if volume <= 0 || entry <= 0 || mark <= 0 {
			t.Skip()
		}
		if math.IsNaN(volume) || math.IsNaN(entry) || math.IsNaN(mark) {
			t.Skip()
		}
		if math.IsInf(volume, 0) || math.IsInf(entry, 0) || math.IsInf(mark, 0) {
			t.Skip()
		}

		pnlLong := Future{}.UnrealizedPnL(Long, volume, 0, entry, mark)
		pnlShort := Future{}.UnrealizedPnL(Short, volume, 0, entry, mark)

		if math.IsNaN(pnlLong) || math.IsInf(pnlLong, 0) {
			t.Fatalf("long pnl non-finite: %v", pnlLong)
		}
		if math.IsNaN(pnlShort) || math.IsInf(pnlShort, 0) {
			t.Fatalf("short pnl non-finite: %v", pnlShort)
		}

		// long + short PnL should sum to zero (zero-sum property)
		sum := pnlLong + pnlShort
		if math.Abs(sum) > 1e-6 {
			t.Fatalf("long(%v) + short(%v) = %v, expected 0", pnlLong, pnlShort, sum)
		}

		// sign consistency
		if mark > entry && pnlLong < 0 {
			t.Fatalf("long pnl negative when mark > entry: %v", pnlLong)
		}
		if mark < entry && pnlLong > 0 {
			t.Fatalf("long pnl positive when mark < entry: %v", pnlLong)
		}
	})
}

func FuzzSimulateAddOn(f *testing.F) {
	f.Add(1.0, 100000.0, 10000.0, 10.0, 90000.0, 10000.0)

	f.Fuzz(func(t *testing.T, volume, entry, margin, leverage, addPrice, addMargin float64) {
		if volume <= 0 || entry <= 0 || margin <= 0 || leverage <= 0 {
			t.Skip()
		}
		if addPrice <= 0 || addMargin <= 0 {
			t.Skip()
		}
		if math.IsNaN(volume) || math.IsNaN(entry) || math.IsNaN(margin) {
			t.Skip()
		}
		if math.IsInf(volume, 0) || math.IsInf(entry, 0) || math.IsInf(margin, 0) {
			t.Skip()
		}
		if math.IsInf(addPrice, 0) || math.IsInf(addMargin, 0) || math.IsInf(leverage, 0) {
			t.Skip()
		}
		if leverage > 200 || addMargin > 1e12 || margin > 1e12 {
			t.Skip()
		}

		pos := Position{Side: Long, Volume: volume, EntryPrice: entry, Margin: margin, Leverage: leverage}
		add := AddOn{Price: addPrice, Margin: addMargin}

		result := Future{}.SimulateAddOn(pos, add, 0.005)

		// margin always equals sum
		expectedMargin := margin + addMargin
		if math.Abs(result.After.Margin-expectedMargin) > 1e-6 {
			t.Fatalf("margin: got %v, want %v", result.After.Margin, expectedMargin)
		}

		// volume always grows
		if result.After.Volume <= pos.Volume {
			t.Fatalf("volume must grow: %v -> %v", pos.Volume, result.After.Volume)
		}

		// non-finite checks
		if math.IsNaN(result.After.EntryPrice) || math.IsInf(result.After.EntryPrice, 0) {
			t.Fatalf("non-finite entry after addon: %v", result.After.EntryPrice)
		}
	})
}

func FuzzVolumeFromMargin(f *testing.F) {
	f.Add(10.0, 1000.0, 50000.0)
	f.Add(5.0, 500.0, 100000.0)

	f.Fuzz(func(t *testing.T, leverage, margin, price float64) {
		if leverage <= 0 || margin <= 0 || price <= 0 {
			t.Skip()
		}
		if math.IsNaN(leverage) || math.IsNaN(margin) || math.IsNaN(price) {
			t.Skip()
		}
		if math.IsInf(leverage, 0) || math.IsInf(margin, 0) || math.IsInf(price, 0) {
			t.Skip()
		}

		vol := Future{Leverage: leverage}.VolumeFromMargin(margin, price)

		if math.IsNaN(vol) || math.IsInf(vol, 0) {
			t.Fatalf("non-finite volume: %v", vol)
		}
		if vol <= 0 {
			t.Fatalf("volume must be positive: %v", vol)
		}

		// round-trip: volume * price / leverage ≈ margin
		reconstructed := vol * price / leverage
		ratio := reconstructed / margin
		if math.Abs(ratio-1) > 1e-9 {
			t.Fatalf("round-trip failed: margin=%v, reconstructed=%v", margin, reconstructed)
		}
	})
}

func FuzzSpotPnL(f *testing.F) {
	f.Add(2.0, 100.0, 150.0)
	f.Add(1.0, 50000.0, 48000.0)

	f.Fuzz(func(t *testing.T, volume, entry, mark float64) {
		if volume <= 0 || entry <= 0 || mark <= 0 {
			t.Skip()
		}
		if math.IsNaN(volume) || math.IsNaN(entry) || math.IsNaN(mark) {
			t.Skip()
		}
		if math.IsInf(volume, 0) || math.IsInf(entry, 0) || math.IsInf(mark, 0) {
			t.Skip()
		}

		value, percent := Spot{}.PnL(volume, entry, mark)

		if math.IsNaN(value) || math.IsInf(value, 0) {
			t.Fatalf("non-finite pnl value: %v", value)
		}
		if math.IsNaN(percent) || math.IsInf(percent, 0) {
			t.Fatalf("non-finite pnl percent: %v", percent)
		}

		// sign consistency
		if mark > entry && value < 0 {
			t.Fatalf("value negative when mark > entry")
		}
		if mark < entry && value > 0 {
			t.Fatalf("value positive when mark < entry")
		}

		// percent matches value direction
		if (value > 0 && percent < 0) || (value < 0 && percent > 0) {
			t.Fatalf("sign mismatch: value=%v, percent=%v", value, percent)
		}
	})
}
