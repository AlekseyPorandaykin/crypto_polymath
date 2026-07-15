package indicator_test

import (
	"context"
	"testing"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator/calculator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/memory"
)

func BenchmarkCalcIndicatorsByCandlestick(b *testing.B) {
	repo := memory.NewIndicatorRepository(10_000)
	dt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	candle := sampleCandle(dt, 100, 110)
	s := indicator.NewService(repo, stubCandlestick{
		lastCandlesticks: func(_ context.Context, _, _ string, _ domain.Unit, _, _ int, _ time.Time) ([]domain.Candlestick, error) {
			return []domain.Candlestick{candle}, nil
		},
	})
	s.AddCalculator(calculator.NewTypeCandle())
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.CalcIndicatorsByCandlestick(ctx, candle, 1); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCalcIndicatorsByCandlestick_cached(b *testing.B) {
	repo := memory.NewIndicatorRepository(10_000)
	dt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	candle := sampleCandle(dt, 100, 110)
	s := indicator.NewService(repo, stubCandlestick{
		lastCandlesticks: func(_ context.Context, _, _ string, _ domain.Unit, _, _ int, _ time.Time) ([]domain.Candlestick, error) {
			return []domain.Candlestick{candle}, nil
		},
	})
	s.AddCalculator(calculator.NewTypeCandle())
	ctx := context.Background()
	if _, err := s.CalcIndicatorsByCandlestick(ctx, candle, 1); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.CalcIndicatorsByCandlestick(ctx, candle, 1); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCalculateLastSequence(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		b.Run(benchSizeLabel(size), func(b *testing.B) {
			benchmarkCalculateLastSequence(b, size)
		})
	}
}

func benchmarkCalculateLastSequence(b *testing.B, size int) {
	repo := memory.NewIndicatorRepository(10_000)
	candles := benchCandles(size)
	s := indicator.NewService(repo, stubCandlestick{
		sequenceCandlesticks: func(_ context.Context, _, _ string, _ domain.Unit, _, _ int) ([]domain.Candlestick, error) {
			return candles, nil
		},
		lastCandlesticks: func(_ context.Context, _, _ string, _ domain.Unit, _, _ int, datetime time.Time) ([]domain.Candlestick, error) {
			for _, candle := range candles {
				if candle.StartTime.Equal(datetime) {
					return []domain.Candlestick{candle}, nil
				}
			}
			return nil, nil
		},
	})
	s.AddCalculator(calculator.NewTypeCandle())
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.CalculateLastSequence(ctx, "bybit", "BTCUSDT", domain.HourUnit, 1, domain.TypeCandleIndicator, 1, size); err != nil {
			b.Fatal(err)
		}
	}
}

func benchCandles(size int) []domain.Candlestick {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]domain.Candlestick, size)
	for i := range candles {
		open := 100 + float64(i%5)
		close := open + float64(i%3) - 1
		candles[i] = sampleCandle(base.Add(time.Duration(i)*time.Hour), open, close)
	}
	return candles
}

func benchSizeLabel(size int) string {
	switch size {
	case 10:
		return "n10"
	case 100:
		return "n100"
	case 1000:
		return "n1000"
	default:
		return "n"
	}
}
