package analysis

import (
	"testing"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/google/uuid"
)

func TestIsCorrectSequence_hourUnit(t *testing.T) {
	base := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	data := []Analytic{
		analyticAt(base, domain.HourUnit, 1),
		analyticAt(base.Add(time.Hour), domain.HourUnit, 1),
		analyticAt(base.Add(2*time.Hour), domain.HourUnit, 1),
	}

	if !isCorrectSequence(data) {
		t.Fatal("expected valid hourly sequence")
	}
}

func TestIsCorrectSequence_detectsGap(t *testing.T) {
	base := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	data := []Analytic{
		analyticAt(base, domain.HourUnit, 1),
		analyticAt(base.Add(2*time.Hour), domain.HourUnit, 1),
	}

	if isCorrectSequence(data) {
		t.Fatal("expected gap to invalidate sequence")
	}
}

func TestIsCorrectSequence_empty(t *testing.T) {
	if isCorrectSequence(nil) {
		t.Fatal("empty sequence must be invalid")
	}
}

func TestIsPrev_minuteUnit(t *testing.T) {
	prev := analyticAt(time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), domain.MinuteUnit, 5)
	current := analyticAt(prev.Datetime.Add(5*time.Minute), domain.MinuteUnit, 5)

	if !isPrev(current, prev) {
		t.Fatal("expected consecutive 5-minute candles")
	}
}

func TestIsPrev_dayUnit(t *testing.T) {
	prev := analyticAt(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), domain.DayUnit, 1)
	current := analyticAt(prev.Datetime.Add(24*time.Hour), domain.DayUnit, 1)

	if !isPrev(current, prev) {
		t.Fatal("expected consecutive daily candles")
	}
}

func analyticAt(dt time.Time, unit domain.Unit, interval int) Analytic {
	return Analytic{
		ID:       uuid.New(),
		Unit:     unit,
		Interval: interval,
		Datetime: dt,
	}
}
