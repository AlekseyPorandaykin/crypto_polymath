package analysis_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/memory"
)

var benchDepths = []int{1, 8, 9, 10, 12, 14, 20, 26, 50}

func BenchmarkCalculateByAnalytics(b *testing.B) {
	for _, size := range []int{1, 10, 100, 1000} {
		b.Run(batchLabel(size), func(b *testing.B) {
			benchmarkCalculateByAnalytics(b, size)
		})
	}
}

func benchmarkCalculateByAnalytics(b *testing.B, size int) {
	s, sources := benchAnalyticService(size)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.CalculateByAnalytics(ctx, sources); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCalculateByAnalytics_cached(b *testing.B) {
	s, sources := benchAnalyticService(100)
	ctx := context.Background()
	if _, err := s.CalculateByAnalytics(ctx, sources); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.CalculateByAnalytics(ctx, sources); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCalculateByAnalytic_perMessage(b *testing.B) {
	for _, size := range []int{1, 10, 100} {
		b.Run(batchLabel(size), func(b *testing.B) {
			benchmarkCalculateByAnalyticPerMessage(b, size)
		})
	}
}

func benchmarkCalculateByAnalyticPerMessage(b *testing.B, size int) {
	s, sources := benchAnalyticService(size)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, source := range sources {
			if _, err := s.CalculateByAnalytic(ctx, source); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkOscillatorByAnalytics(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		b.Run(batchLabel(size), func(b *testing.B) {
			benchmarkOscillatorByAnalytics(b, size)
		})
	}
}

func benchmarkOscillatorByAnalytics(b *testing.B, size int) {
	s, sources := benchAnalyticService(size)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.OscillatorByAnalytics(ctx, sources, "derived", 1); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCalculateByIndicator(b *testing.B) {
	repo := memory.NewAnalysisRepository(10_000)
	s := analysis.NewService(repo, nil, benchDepths)
	s.AddCalculatorByIndicator(stubCalculatorByIndicator{
		name:        "analytic",
		byIndicator: domain.EMAIndicator,
		depths:      map[int]bool{1: true},
		onCalculate: func(_ context.Context, indicator domain.Indicator, depth int) (*analysis.Analytic, error) {
			return &analysis.Analytic{
				Name:           "analytic",
				Exchange:       indicator.Exchange,
				Symbol:         indicator.Symbol,
				Unit:           indicator.Unit,
				Interval:       indicator.Interval,
				Datetime:       indicator.Datetime,
				Depth:          depth,
				ByIndicator:    indicator.Name,
				IndicatorDepth: indicator.Depth,
				Value:          indicator.Value + 1,
			}, nil
		},
	})

	indicatorData := domain.Indicator{
		Exchange: "bybit",
		Symbol:   "BTCUSDT",
		Unit:     domain.HourUnit,
		Interval: 1,
		Datetime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		Name:     domain.EMAIndicator,
		Depth:    26,
		Value:    100,
	}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.CalculateByIndicator(ctx, indicatorData); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAnalyticByIndicators(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		b.Run(batchLabel(size), func(b *testing.B) {
			benchmarkAnalyticByIndicators(b, size)
		})
	}
}

func benchmarkAnalyticByIndicators(b *testing.B, size int) {
	repo := memory.NewAnalysisRepository(10_000)
	s := analysis.NewService(repo, nil, []int{1})
	s.AddCalculatorByIndicator(stubCalculatorByIndicator{
		name:        "analytic",
		byIndicator: domain.EMAIndicator,
		depths:      map[int]bool{1: true},
		onCalculate: func(_ context.Context, indicator domain.Indicator, depth int) (*analysis.Analytic, error) {
			return &analysis.Analytic{
				Name:           "analytic",
				Exchange:       indicator.Exchange,
				Symbol:         indicator.Symbol,
				Unit:           indicator.Unit,
				Interval:       indicator.Interval,
				Datetime:       indicator.Datetime,
				Depth:          depth,
				ByIndicator:    indicator.Name,
				IndicatorDepth: indicator.Depth,
				Value:          float64(indicator.Datetime.Unix()),
			}, nil
		},
	})

	indicators := benchIndicators(size)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.AnalyticByIndicators(ctx, indicators, "analytic", 1); err != nil {
			b.Fatal(err)
		}
	}
}

func benchAnalyticService(size int) (*analysis.Service, []analysis.Analytic) {
	s := analysis.NewService(memory.NewAnalysisRepository(10_000), nil, benchDepths)
	s.AddCalculatorByAnalytic(stubCalculatorByAnalytic{
		name:       "derived",
		byAnalytic: "source",
		depths:     map[int]bool{1: true},
		onCalculate: func(_ context.Context, data analysis.Analytic, depth int) (*analysis.Analytic, error) {
			return derivedAnalytic(data, depth, data.Value*2), nil
		},
	})
	return s, benchSourceAnalytics(size)
}

func benchSourceAnalytics(size int) []analysis.Analytic {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	sources := make([]analysis.Analytic, size)
	for i := range sources {
		sources[i] = sampleAnalyticWithValue("source", base.Add(time.Duration(i)*time.Hour), 1, float64(i+1))
	}
	return sources
}

func benchIndicators(size int) []domain.Indicator {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	indicators := make([]domain.Indicator, size)
	for i := range indicators {
		indicators[i] = domain.Indicator{
			Exchange: "bybit",
			Symbol:   "BTCUSDT",
			Unit:     domain.HourUnit,
			Interval: 1,
			Datetime: base.Add(time.Duration(i) * time.Hour),
			Name:     domain.EMAIndicator,
			Depth:    26,
			Value:    100,
		}
	}
	return indicators
}

func batchLabel(size int) string {
	switch size {
	case 1:
		return "n1"
	case 10:
		return "n10"
	case 100:
		return "n100"
	case 1000:
		return "n1000"
	default:
		return "n" + strconv.Itoa(size)
	}
}
