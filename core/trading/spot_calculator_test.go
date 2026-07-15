package trading

import (
	"math"
	"testing"
)

func TestSpot_PnL(t *testing.T) {
	cases := []struct {
		name        string
		volume      float64
		entry       float64
		mark        float64
		wantValue   float64
		wantPercent float64
	}{
		{name: "profit", volume: 2, entry: 100, mark: 150, wantValue: 100, wantPercent: 50},
		{name: "loss", volume: 2, entry: 100, mark: 80, wantValue: -40, wantPercent: -20},
		{name: "no change", volume: 1, entry: 100, mark: 100, wantValue: 0, wantPercent: 0},
		{name: "zero entry", volume: 1, entry: 0, mark: 100, wantValue: 100, wantPercent: 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			value, percent := (Spot{}).PnL(c.volume, c.entry, c.mark)
			if math.Abs(value-c.wantValue) > 1e-9 {
				t.Fatalf("value: got %v, want %v", value, c.wantValue)
			}
			if math.Abs(percent-c.wantPercent) > 1e-9 {
				t.Fatalf("percent: got %v, want %v", percent, c.wantPercent)
			}
		})
	}
}
