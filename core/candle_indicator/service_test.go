package candle_indicator_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/candle_indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/candlestick"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/memory"
)

func TestService_CalculateFromCandlesticks_computesBatch(t *testing.T) {
	repo := memory.NewCandleIndicatorRepository(100)
	s := candle_indicator.New(repo, stubCandlestickService{})
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

	t1 := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
	candles := []domain.Candlestick{
		sampleCandle(t1, 100, 110),
		sampleCandle(t2, 110, 105),
	}

	result, err := s.CalculateFromCandlesticks(context.Background(), candles)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 indicators, got %#v", result)
	}

	stored, err := repo.FindMany(context.Background(), "TestIndicator", "bybit", "BTCUSDT", string(domain.HourUnit), 1, []time.Time{t1, t2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stored) != 2 {
		t.Fatalf("expected 2 stored rows, got %#v", stored)
	}
}

func TestService_CalculateFromCandlesticks_returnsCached(t *testing.T) {
	repo := memory.NewCandleIndicatorRepository(100)
	t1 := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	if err := repo.Save(context.Background(), []candle_indicator.StorageDTO{
		{
			Name: "TestIndicator", Exchange: "bybit", Symbol: "BTCUSDT", Unit: string(domain.HourUnit),
			Interval: 1, StartTime: t1, OpenPrice: 1, ClosePrice: 2, HighPrice: 3, LowPrice: 0,
		},
	}); err != nil {
		t.Fatalf("save fixture: %v", err)
	}

	var calls atomic.Int32
	s := candle_indicator.New(repo, stubCandlestickService{})
	s.AddCalculator(stubCalculator{
		name: "TestIndicator",
		calculate: func(_ context.Context, _ domain.Candlestick) (*candle_indicator.Indicator, error) {
			calls.Add(1)
			return nil, nil
		},
	})

	result, err := s.CalculateFromCandlesticks(context.Background(), []domain.Candlestick{sampleCandle(t1, 100, 110)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].ClosePrice != 2 {
		t.Fatalf("expected cached indicator, got %#v", result)
	}
	if calls.Load() != 0 {
		t.Fatalf("calculator must not be called, calls=%d", calls.Load())
	}
}

func TestService_CalculateFromCandlesticks_emptyInput(t *testing.T) {
	s := candle_indicator.New(memory.NewCandleIndicatorRepository(10), stubCandlestickService{})
	result, err := s.CalculateFromCandlesticks(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil, got %#v", result)
	}
}

type stubCalculator struct {
	name      string
	calculate func(ctx context.Context, candle domain.Candlestick) (*candle_indicator.Indicator, error)
}

func (s stubCalculator) Name() string { return s.name }
func (s stubCalculator) Calculate(ctx context.Context, candle domain.Candlestick) (*candle_indicator.Indicator, error) {
	if s.calculate == nil {
		return nil, nil
	}
	return s.calculate(ctx, candle)
}

type stubCandlestickService struct{}

func (stubCandlestickService) AddLoader(string, candlestick.ExchangeLoader) {}
func (stubCandlestickService) LoadCandlesticks(context.Context, string, string, domain.Unit, int) ([]domain.Candlestick, error) {
	return nil, nil
}
func (stubCandlestickService) DeleteOldRows(context.Context, int) error { return nil }
func (stubCandlestickService) CandlesticksToDate(context.Context, string, string, string, int, int, time.Time) ([]domain.Candlestick, error) {
	return nil, nil
}
func (stubCandlestickService) UpdateCandlesticks(context.Context, string, string, domain.Unit, int) ([]domain.Candlestick, error) {
	return nil, nil
}
func (stubCandlestickService) SequenceCandlesticksToDate(context.Context, string, string, string, int, int, time.Time) ([]domain.Candlestick, error) {
	return nil, nil
}
func (stubCandlestickService) CandlesticksFromDate(context.Context, string, string, string, int, int, time.Time) ([]domain.Candlestick, error) {
	return nil, nil
}
func (stubCandlestickService) SequenceCandlesticks(context.Context, string, string, domain.Unit, int, int) ([]domain.Candlestick, error) {
	return nil, nil
}
func (stubCandlestickService) Candlesticks(context.Context, string, string, domain.Unit, int, int) ([]domain.Candlestick, error) {
	return nil, nil
}

func sampleCandle(dt time.Time, open, close float64) domain.Candlestick {
	high, low := open, close
	if close > high {
		high = close
	}
	if close < low {
		low = close
	}
	return domain.Candlestick{
		Exchange: "bybit", Symbol: "BTCUSDT", Unit: domain.HourUnit, Interval: 1,
		StartTime: dt, OpenPrice: open, ClosePrice: close, HighPrice: high, LowPrice: low, Volume: 1,
	}
}
