package candle_indicator

import (
	"testing"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

func TestIndicator_SizeBodyAndPercent(t *testing.T) {
	ind := Indicator{
		OpenPrice: 100, ClosePrice: 110, HighPrice: 120, LowPrice: 90,
	}
	if ind.SizeBody() != 10 {
		t.Fatalf("expected body size 10, got %v", ind.SizeBody())
	}
	if ind.Size() != 30 {
		t.Fatalf("expected total size 30, got %v", ind.Size())
	}
	wantPercent := 10.0 / 30.0 * 100
	gotPercent := ind.SizeBodyInPercent()
	if gotPercent < wantPercent-0.0001 || gotPercent > wantPercent+0.0001 {
		t.Fatalf("expected ~%.2f%%, got %v", wantPercent, gotPercent)
	}
}

func TestIndicator_SizeBodyInPercent_zeroBody(t *testing.T) {
	ind := Indicator{OpenPrice: 100, ClosePrice: 100, HighPrice: 110, LowPrice: 90}
	if ind.SizeBodyInPercent() != 0 {
		t.Fatalf("expected 0 for doji, got %v", ind.SizeBodyInPercent())
	}
}

func TestIndicator_directionHelpers(t *testing.T) {
	up := Indicator{OpenPrice: 100, ClosePrice: 110}
	if !up.IsUp() || up.IsDown() || up.Direction() != domain.UpDirection {
		t.Fatalf("unexpected up candle helpers: %#v", up.Direction())
	}

	down := Indicator{OpenPrice: 110, ClosePrice: 100}
	if !down.IsDown() || down.IsUp() || down.Direction() != domain.DownDirection {
		t.Fatalf("unexpected down candle helpers: %#v", down.Direction())
	}

	flat := Indicator{OpenPrice: 100, ClosePrice: 100}
	if flat.IsUp() || flat.IsDown() || flat.Direction() != domain.IndefiniteDirection {
		t.Fatalf("unexpected flat candle helpers: %#v", flat.Direction())
	}
}

func TestIndicator_PrevStartTime(t *testing.T) {
	start := time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)
	ind := Indicator{Unit: domain.HourUnit, Interval: 1, StartTime: start}
	prev := ind.PrevStartTime()
	want := start.Add(-time.Hour)
	if !prev.Equal(want) {
		t.Fatalf("expected %v, got %v", want, prev)
	}
}
