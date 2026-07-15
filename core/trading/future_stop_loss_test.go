package trading

import (
	"math"
	"testing"
)

func TestVolatilityRangePercent(t *testing.T) {
	got := VolatilityRangePercent(105, 95, 100)
	if math.Abs(got-10) > 1e-9 {
		t.Fatalf("got %v, want 10", got)
	}
}

func TestIsHeikenAshiDoji(t *testing.T) {
	if !IsHeikenAshiDoji(100, 100, 110, 90, 10) {
		t.Fatal("zero body must be doji")
	}
	if !IsHeikenAshiDoji(100, 101, 110, 90, 10) {
		t.Fatal("1/20 body should be doji at 10% threshold")
	}
	if IsHeikenAshiDoji(100, 105, 110, 90, 10) {
		t.Fatal("5/20 body should not be doji at 10% threshold")
	}
}

func TestFuture_MarginPnLPercent_and_PriceForMarginPnLPercent_long(t *testing.T) {
	f := Future{Leverage: 5}
	entry := 60_000.0

	marginPct := f.MarginPnLPercent(Long, entry, 61_200) // +2% price
	if math.Abs(marginPct-10) > 1e-9 {
		t.Fatalf("margin pnl: got %v, want 10", marginPct)
	}

	stop := f.PriceForMarginPnLPercent(Long, entry, -10)
	wantStop := 58_800.0 // -2% price at 5x
	if math.Abs(stop-wantStop) > 1e-6 {
		t.Fatalf("stop price: got %v, want %v", stop, wantStop)
	}
}

func TestFuture_MarginPnLPercent_short(t *testing.T) {
	f := Future{Leverage: 3}
	entry := 50_000.0
	marginPct := f.MarginPnLPercent(Short, entry, 49_000) // +2.04% approx
	if marginPct <= 0 {
		t.Fatalf("short profit must be positive, got %v", marginPct)
	}
}

func TestFuture_DynamicStopLoss_range(t *testing.T) {
	f := Future{Leverage: 5}
	coef := DefaultStopLossCoefficients()
	vol := VolatilitySnapshot{RangePct: 1.5, ATRPct: 1.2, MktVol: 1.3}
	entry := 62_000.0

	sl := f.DynamicStopLoss(Long, entry, vol, coef, VolatilityRange)

	if math.Abs(sl.SLPricePct-6.0) > 1e-9 {
		t.Fatalf("sl price pct: got %v, want 6", sl.SLPricePct)
	}
	if math.Abs(sl.TrailActivatePricePct-6.0) > 1e-9 {
		t.Fatalf("trail pct: got %v, want 6", sl.TrailActivatePricePct)
	}
	if math.Abs(sl.InitialSLMarginPct+30) > 1e-9 {
		t.Fatalf("initial sl margin: got %v, want -30", sl.InitialSLMarginPct)
	}
	if math.Abs(sl.TrailActivateMarginPct-30) > 1e-9 {
		t.Fatalf("trail margin: got %v, want 30", sl.TrailActivateMarginPct)
	}

	wantStop := entry * (1 - 0.06)
	if math.Abs(sl.InitialStopPrice-wantStop) > 1e-6 {
		t.Fatalf("initial stop: got %v, want %v", sl.InitialStopPrice, wantStop)
	}
}

func TestFuture_DynamicStopLoss_floor(t *testing.T) {
	f := Future{Leverage: 5}
	coef := DefaultStopLossCoefficients()
	vol := VolatilitySnapshot{RangePct: 0.2} // very low vol
	entry := 100.0

	sl := f.DynamicStopLoss(Long, entry, vol, coef, VolatilityRange)
	if math.Abs(sl.SLPricePct-coef.SLFloorPct) > 1e-9 {
		t.Fatalf("expected floor %v, got %v", coef.SLFloorPct, sl.SLPricePct)
	}
}

func TestFuture_UpdateTrailingStop_long(t *testing.T) {
	f := Future{Leverage: 5}
	upd := f.UpdateTrailingStop(Long, 100, 99, 101, 110, true)
	if upd.StopPrice != 101 {
		t.Fatalf("expected stop 101, got %v", upd.StopPrice)
	}
	if !upd.TrailActivated {
		t.Fatal("trail must stay activated")
	}

	upd2 := f.UpdateTrailingStop(Long, 100, 99, 98, 110, false)
	if upd2.StopPrice != 99 {
		t.Fatalf("before activation stop unchanged: got %v", upd2.StopPrice)
	}
}

func TestFuture_UpdateTrailingStop_short(t *testing.T) {
	f := Future{Leverage: 5}
	upd := f.UpdateTrailingStop(Short, 100, 101, 90, 99, true)
	if upd.StopPrice != 99 {
		t.Fatalf("expected stop 99, got %v", upd.StopPrice)
	}
}

func TestFuture_ShouldActivateTrailing(t *testing.T) {
	f := Future{Leverage: 5}
	if f.ShouldActivateTrailing(29.9, 30) {
		t.Fatal("should not activate below threshold")
	}
	if !f.ShouldActivateTrailing(30, 30) {
		t.Fatal("should activate at threshold")
	}
}

func TestIsStopHit_and_StopExitPrice(t *testing.T) {
	if !IsStopHit(Long, 100, 99, 105) {
		t.Fatal("long stop hit when low <= stop")
	}
	if IsStopHit(Long, 100, 101, 105) {
		t.Fatal("long stop not hit when low > stop")
	}
	if StopExitPrice(Long, 100, 98) != 98 {
		t.Fatalf("gap down exit at open")
	}
	if StopExitPrice(Long, 100, 101) != 100 {
		t.Fatalf("normal exit at stop")
	}
}

func TestDefaultStopLossCoefficients(t *testing.T) {
	c := DefaultStopLossCoefficients()
	if c.KSL != 4.0 || c.KTrail != 4.0 || c.DojiBodyPct != 10.0 {
		t.Fatalf("unexpected defaults: %#v", c)
	}
}
