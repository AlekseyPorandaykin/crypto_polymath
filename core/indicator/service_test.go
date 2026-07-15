package indicator_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator/calculator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/memory"
	"github.com/google/uuid"
)

func TestService_CalcIndicatorsByCandlestick_computesAndSaves(t *testing.T) {
	repo := memory.NewIndicatorRepository(100)
	dt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	candle := sampleCandle(dt, 100, 110)

	s := indicator.NewService(repo, stubCandlestick{
		lastCandlesticks: func(_ context.Context, _, _ string, _ domain.Unit, _, _ int, _ time.Time) ([]domain.Candlestick, error) {
			return []domain.Candlestick{candle}, nil
		},
	})
	s.AddCalculator(calculator.NewTypeCandle())

	result, err := s.CalcIndicatorsByCandlestick(context.Background(), candle, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 indicator, got %#v", result)
	}
	if result[0].Name != domain.TypeCandleIndicator || result[0].Value != float64(domain.UpCandle) {
		t.Fatalf("unexpected indicator: %#v", result[0])
	}

	stored, err := repo.Find(context.Background(), candle.Exchange, candle.Symbol, string(candle.Unit), candle.Interval, dt, domain.TypeCandleIndicator, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stored == nil {
		t.Fatal("expected indicator in storage")
	}
}

func TestService_CalcIndicatorsByCandlestick_returnsCached(t *testing.T) {
	repo := memory.NewIndicatorRepository(100)
	dt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	candle := sampleCandle(dt, 100, 110)
	if err := repo.Save(context.Background(), indicator.StorageDTO{
		ID: uuid.New(), Exchange: candle.Exchange, Symbol: candle.Symbol, Unit: string(candle.Unit),
		Interval: candle.Interval, Name: domain.TypeCandleIndicator, Datetime: dt, Depth: 1, Value: 999,
	}); err != nil {
		t.Fatalf("save fixture: %v", err)
	}

	var calls atomic.Int32
	s := indicator.NewService(repo, stubCandlestick{
		lastCandlesticks: func(_ context.Context, _, _ string, _ domain.Unit, _, _ int, _ time.Time) ([]domain.Candlestick, error) {
			calls.Add(1)
			return []domain.Candlestick{candle}, nil
		},
	})
	s.AddCalculator(calculator.NewTypeCandle())

	result, err := s.CalcIndicatorsByCandlestick(context.Background(), candle, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Value != 999 {
		t.Fatalf("expected cached value 999, got %#v", result)
	}
	if calls.Load() != 0 {
		t.Fatalf("candlestick provider must not be called when cache hit, calls=%d", calls.Load())
	}
}

func TestService_CalcIndicatorsByCandlestick_emptyDepth(t *testing.T) {
	s := indicator.NewService(memory.NewIndicatorRepository(10), stubCandlestick{})
	s.AddCalculator(calculator.NewTypeCandle())

	result, err := s.CalcIndicatorsByCandlestick(context.Background(), sampleCandle(time.Now(), 1, 2), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil for depth < 1, got %#v", result)
	}
}

func TestService_CalculateLastSequence_batchProcessing(t *testing.T) {
	repo := memory.NewIndicatorRepository(100)
	base := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
	candles := []domain.Candlestick{
		sampleCandle(base, 100, 105),
		sampleCandle(base.Add(time.Hour), 105, 103),
	}

	s := indicator.NewService(repo, stubCandlestick{
		sequenceCandlesticks: func(_ context.Context, _, _ string, _ domain.Unit, _, _ int) ([]domain.Candlestick, error) {
			return candles, nil
		},
		lastCandlesticks: func(_ context.Context, _, _ string, _ domain.Unit, _, limit int, datetime time.Time) ([]domain.Candlestick, error) {
			for _, candle := range candles {
				if candle.StartTime.Equal(datetime) {
					return []domain.Candlestick{candle}, nil
				}
			}
			return nil, nil
		},
	})
	s.AddCalculator(calculator.NewTypeCandle())

	result, err := s.CalculateLastSequence(context.Background(), "bybit", "BTCUSDT", domain.HourUnit, 1, domain.TypeCandleIndicator, 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 indicators, got %#v", result)
	}
	if result[0].Value != float64(domain.UpCandle) {
		t.Fatalf("first candle expected up, got %#v", result[0])
	}
	if result[1].Value != float64(domain.DownCandle) {
		t.Fatalf("second candle expected down, got %#v", result[1])
	}
}

func TestService_Indicators_readsFromRepo(t *testing.T) {
	repo := memory.NewIndicatorRepository(100)
	dt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	if err := repo.Save(context.Background(), indicator.StorageDTO{
		ID: uuid.New(), Exchange: "bybit", Symbol: "BTCUSDT", Unit: string(domain.HourUnit),
		Interval: 1, Name: domain.TypeCandleIndicator, Datetime: dt, Depth: 1, Value: 42,
	}); err != nil {
		t.Fatalf("save fixture: %v", err)
	}

	s := indicator.NewService(repo, stubCandlestick{})
	result, err := s.Indicators(context.Background(), "bybit", "BTCUSDT", domain.HourUnit, 1, domain.TypeCandleIndicator, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Value != 42 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

type stubCandlestick struct {
	lastCandlesticks     func(ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int, datetime time.Time) ([]domain.Candlestick, error)
	nextCandlesticks     func(ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int, datetime time.Time) ([]domain.Candlestick, error)
	firstCandlestick     func(ctx context.Context, exchange, symbol string, unit domain.Unit, interval, offset int) (*domain.Candlestick, error)
	sequenceCandlesticks func(ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int) ([]domain.Candlestick, error)
}

func (s stubCandlestick) LastCandlesticks(ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int, datetime time.Time) ([]domain.Candlestick, error) {
	if s.lastCandlesticks != nil {
		return s.lastCandlesticks(ctx, exchange, symbol, unit, interval, limit, datetime)
	}
	return nil, nil
}

func (s stubCandlestick) NextCandlesticks(ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int, datetime time.Time) ([]domain.Candlestick, error) {
	if s.nextCandlesticks != nil {
		return s.nextCandlesticks(ctx, exchange, symbol, unit, interval, limit, datetime)
	}
	return nil, nil
}

func (s stubCandlestick) FirstCandlestick(ctx context.Context, exchange, symbol string, unit domain.Unit, interval, offset int) (*domain.Candlestick, error) {
	if s.firstCandlestick != nil {
		return s.firstCandlestick(ctx, exchange, symbol, unit, interval, offset)
	}
	return nil, nil
}

func (s stubCandlestick) SequenceCandlesticks(ctx context.Context, exchange, symbol string, unit domain.Unit, interval, limit int) ([]domain.Candlestick, error) {
	if s.sequenceCandlesticks != nil {
		return s.sequenceCandlesticks(ctx, exchange, symbol, unit, interval, limit)
	}
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
		Exchange:   "bybit",
		Symbol:     "BTCUSDT",
		Unit:       domain.HourUnit,
		Interval:   1,
		StartTime:  dt,
		OpenPrice:  open,
		ClosePrice: close,
		HighPrice:  high,
		LowPrice:   low,
		Volume:     1,
	}
}
