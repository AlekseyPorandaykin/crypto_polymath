package candle_indicator_test

import (
	"context"
	"testing"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/memory"
)

func BenchmarkCalculateFromCandlesticks(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		b.Run(benchSizeLabel(size), func(b *testing.B) {
			benchmarkCalculateFromCandlesticks(b, size)
		})
	}
}

func benchmarkCalculateFromCandlesticks(b *testing.B, size int) {
	repo := memory.NewCandleIndicatorRepository(10_000)
	s := candle_indicator.New(repo, benchCandlestickService{})
	s.AddCalculator(stubCalculator{
		name: "TestIndicator",
		calculate: func(_ context.Context, candle domain.Candlestick) (*candle_indicator.Indicator, error) {
			return &candle_indicator.Indicator{
				Name:       "TestIndicator",
				Exchange:   candle.Exchange,
				Symbol:     candle.Symbol,
				Unit:       candle.Unit,
				Interval:   candle.Interval,
				StartTime:  candle.StartTime,
				OpenPrice:  candle.OpenPrice,
				ClosePrice: candle.ClosePrice,
				HighPrice:  candle.HighPrice,
				LowPrice:   candle.LowPrice,
			}, nil
		},
	})

	candles := benchCandles(size)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.CalculateFromCandlesticks(ctx, candles); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCalculateFromCandlesticks_cached(b *testing.B) {
	repo := memory.NewCandleIndicatorRepository(10_000)
	s := candle_indicator.New(repo, benchCandlestickService{})
	s.AddCalculator(stubCalculator{
		name: "TestIndicator",
		calculate: func(_ context.Context, candle domain.Candlestick) (*candle_indicator.Indicator, error) {
			return &candle_indicator.Indicator{
				Name: "TestIndicator", Exchange: candle.Exchange, Symbol: candle.Symbol,
				Unit: candle.Unit, Interval: candle.Interval, StartTime: candle.StartTime,
				OpenPrice: candle.OpenPrice, ClosePrice: candle.ClosePrice,
				HighPrice: candle.HighPrice, LowPrice: candle.LowPrice,
			}, nil
		},
	})

	candles := benchCandles(100)
	ctx := context.Background()
	if _, err := s.CalculateFromCandlesticks(ctx, candles); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.CalculateFromCandlesticks(ctx, candles); err != nil {
			b.Fatal(err)
		}
	}
}

func benchCandles(size int) []domain.Candlestick {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]domain.Candlestick, size)
	for i := range candles {
		open := 100 + float64(i%7)
		close := open + float64(i%4) - 1.5
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

// stubCandlestickService is duplicated minimally for bench file compile isolation.
type benchCandlestickService struct{}

func (benchCandlestickService) AddLoader(string, candlestick.ExchangeLoader) {}
func (benchCandlestickService) LoadCandlesticks(context.Context, string, string, domain.Unit, int) ([]domain.Candlestick, error) {
	return nil, nil
}
func (benchCandlestickService) DeleteOldRows(context.Context, int) error { return nil }
func (benchCandlestickService) CandlesticksToDate(context.Context, string, string, string, int, int, time.Time) ([]domain.Candlestick, error) {
	return nil, nil
}
func (benchCandlestickService) UpdateCandlesticks(context.Context, string, string, domain.Unit, int) ([]domain.Candlestick, error) {
	return nil, nil
}
func (benchCandlestickService) SequenceCandlesticksToDate(context.Context, string, string, string, int, int, time.Time) ([]domain.Candlestick, error) {
	return nil, nil
}
func (benchCandlestickService) CandlesticksFromDate(context.Context, string, string, string, int, int, time.Time) ([]domain.Candlestick, error) {
	return nil, nil
}
func (benchCandlestickService) SequenceCandlesticks(context.Context, string, string, domain.Unit, int, int) ([]domain.Candlestick, error) {
	return nil, nil
}
func (benchCandlestickService) Candlesticks(context.Context, string, string, domain.Unit, int, int) ([]domain.Candlestick, error) {
	return nil, nil
}
