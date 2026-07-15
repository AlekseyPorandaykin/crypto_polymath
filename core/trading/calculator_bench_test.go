package trading

import "testing"

func BenchmarkNewAvgEntryPrice(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Future{}.NewAvgEntryPrice(3, 100, 1, 200)
	}
}

func BenchmarkSimulateAddOn_long(b *testing.B) {
	pos := Position{Side: Long, Volume: 1, EntryPrice: 100_000, Margin: 10_000, Leverage: 10}
	add := AddOn{Price: 90_000, Margin: 10_000}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Future{}.SimulateAddOn(pos, add, 0.005)
	}
}

func BenchmarkRiskAtPrice(b *testing.B) {
	pos := Position{Side: Long, Volume: 1, EntryPrice: 100_000, Margin: 10_000, Leverage: 10}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Future{}.RiskAtPrice(pos, 95_000, 0.005)
	}
}
