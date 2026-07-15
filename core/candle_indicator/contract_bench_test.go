package candle_indicator

import "testing"

func BenchmarkIndicator_SizeBodyInPercent(b *testing.B) {
	ind := Indicator{
		OpenPrice: 100, ClosePrice: 110, HighPrice: 120, LowPrice: 90,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ind.SizeBodyInPercent()
	}
}

func BenchmarkIndicator_Direction(b *testing.B) {
	ind := Indicator{OpenPrice: 100, ClosePrice: 110}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ind.Direction()
	}
}
