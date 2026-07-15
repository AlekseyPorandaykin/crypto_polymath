package analysis

import (
	"testing"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/google/uuid"
)

func BenchmarkIsCorrectSequence(b *testing.B) {
	data := benchHourlySequence(100)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !isCorrectSequence(data) {
			b.Fatal("expected valid sequence")
		}
	}
}

func benchHourlySequence(size int) []Analytic {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	data := make([]Analytic, size)
	for i := range data {
		data[i] = Analytic{
			ID:       uuid.New(),
			Unit:     domain.HourUnit,
			Interval: 1,
			Datetime: base.Add(time.Duration(i) * time.Hour),
		}
	}
	return data
}
