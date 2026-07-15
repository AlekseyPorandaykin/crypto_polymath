package analysis_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
	"github.com/AlekseyPorandaykin/crypto_polymath/internal/infrastructure/memory"
	"github.com/google/uuid"
)

func TestService_CalculateByAnalytics_emptyInput(t *testing.T) {
	s := analysis.NewService(memory.NewAnalysisRepository(10), nil, []int{1, 8})
	s.AddCalculatorByAnalytic(stubCalculatorByAnalytic{
		name:       "derived",
		byAnalytic: "source",
	})

	result, err := s.CalculateByAnalytics(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

func TestService_CalculateByAnalytics_noCalculators(t *testing.T) {
	s := analysis.NewService(memory.NewAnalysisRepository(10), nil, []int{1})
	source := sampleAnalytic("source", time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), 1)

	result, err := s.CalculateByAnalytics(context.Background(), []analysis.Analytic{source})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

func TestService_CalculateByAnalytics_skipsUnrelatedSourceName(t *testing.T) {
	var calls atomic.Int32
	s := newServiceWithAnalyticCalculator(t, stubCalculatorByAnalytic{
		name:       "derived",
		byAnalytic: "MACDMainLine",
		depths:     map[int]bool{1: true},
		onCalculate: func(_ context.Context, _ analysis.Analytic, _ int) (*analysis.Analytic, error) {
			calls.Add(1)
			return nil, nil
		},
	})
	source := sampleAnalytic("RSI", time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), 1)

	result, err := s.CalculateByAnalytics(context.Background(), []analysis.Analytic{source})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %#v", result)
	}
	if calls.Load() != 0 {
		t.Fatalf("calculator must not be called, calls=%d", calls.Load())
	}
}

func TestService_CalculateByAnalytics_skipsUnsupportedDepth(t *testing.T) {
	var calls atomic.Int32
	s := newServiceWithAnalyticCalculator(t, stubCalculatorByAnalytic{
		name:       "derived",
		byAnalytic: "source",
		depths:     map[int]bool{1: true},
		onCalculate: func(_ context.Context, data analysis.Analytic, depth int) (*analysis.Analytic, error) {
			calls.Add(1)
			return derivedAnalytic(data, depth, 42), nil
		},
	})
	source := sampleAnalytic("source", time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), 1)

	result, err := s.CalculateByAnalytics(context.Background(), []analysis.Analytic{source})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %#v", result)
	}
	if result[0].Depth != 1 {
		t.Fatalf("expected depth=1, got %d", result[0].Depth)
	}
	if calls.Load() != 1 {
		t.Fatalf("calculator must run once, calls=%d", calls.Load())
	}
}

func TestService_CalculateByAnalytics_computesBatchAndSaves(t *testing.T) {
	repo := memory.NewAnalysisRepository(100)
	s := analysis.NewService(repo, nil, []int{1, 8, 9})
	s.AddCalculatorByAnalytic(stubCalculatorByAnalytic{
		name:       "derived",
		byAnalytic: "source",
		depths:     map[int]bool{1: true},
		onCalculate: func(_ context.Context, data analysis.Analytic, depth int) (*analysis.Analytic, error) {
			return derivedAnalytic(data, depth, data.Value*2), nil
		},
	})

	t1 := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
	sources := []analysis.Analytic{
		sampleAnalyticWithValue("source", t1, 1, 10),
		sampleAnalyticWithValue("source", t2, 1, 20),
	}

	result, err := s.CalculateByAnalytics(context.Background(), sources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %#v", result)
	}
	if result[0].Value != 20 || result[1].Value != 40 {
		t.Fatalf("unexpected values: %#v", result)
	}

	stored, err := repo.FindMany(context.Background(), "derived", "bybit", "BTCUSDT", domain.HourUnit, 1, 1, 1, []time.Time{t1, t2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stored) != 2 {
		t.Fatalf("expected 2 stored rows, got %#v", stored)
	}
}

func TestService_CalculateByAnalytics_returnsExistingWithoutRecalculation(t *testing.T) {
	repo := memory.NewAnalysisRepository(100)
	t1 := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	existing := derivedAnalytic(sampleAnalyticWithValue("source", t1, 1, 10), 1, 99)
	if err := repo.Save(context.Background(), *existing); err != nil {
		t.Fatalf("save fixture: %v", err)
	}

	var calls atomic.Int32
	s := analysis.NewService(repo, nil, []int{1})
	s.AddCalculatorByAnalytic(stubCalculatorByAnalytic{
		name:       "derived",
		byAnalytic: "source",
		depths:     map[int]bool{1: true},
		onCalculate: func(_ context.Context, _ analysis.Analytic, _ int) (*analysis.Analytic, error) {
			calls.Add(1)
			return nil, nil
		},
	})

	result, err := s.CalculateByAnalytics(context.Background(), []analysis.Analytic{
		sampleAnalyticWithValue("source", t1, 1, 10),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Value != 99 {
		t.Fatalf("expected cached value 99, got %#v", result)
	}
	if calls.Load() != 0 {
		t.Fatalf("calculator must not be called, calls=%d", calls.Load())
	}
}

func TestService_CalculateByAnalytic_delegatesToBatch(t *testing.T) {
	s := newServiceWithAnalyticCalculator(t, stubCalculatorByAnalytic{
		name:       "derived",
		byAnalytic: "source",
		depths:     map[int]bool{1: true},
		onCalculate: func(_ context.Context, data analysis.Analytic, depth int) (*analysis.Analytic, error) {
			return derivedAnalytic(data, depth, 5), nil
		},
	})
	source := sampleAnalyticWithValue("source", time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), 1, 3)

	result, err := s.CalculateByAnalytic(context.Background(), source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Value != 5 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestService_OscillatorByAnalytics_batchProcessing(t *testing.T) {
	repo := memory.NewAnalysisRepository(100)
	s := analysis.NewService(repo, nil, []int{1})
	s.AddCalculatorByAnalytic(stubCalculatorByAnalytic{
		name:       "derived",
		byAnalytic: "source",
		depths:     map[int]bool{1: true},
		onCalculate: func(_ context.Context, data analysis.Analytic, depth int) (*analysis.Analytic, error) {
			return derivedAnalytic(data, depth, float64(data.Datetime.Hour())), nil
		},
	})

	times := []time.Time{
		time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	sources := make([]analysis.Analytic, len(times))
	for i, dt := range times {
		sources[i] = sampleAnalyticWithValue("source", dt, 1, 1)
	}

	result, err := s.OscillatorByAnalytics(context.Background(), sources, "derived", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %#v", result)
	}
	if result[0].Value != 10 || result[1].Value != 11 || result[2].Value != 12 {
		t.Fatalf("unexpected values: %#v", result)
	}
}

func TestService_CalculateByIndicator_computesAndSaves(t *testing.T) {
	repo := memory.NewAnalysisRepository(100)
	s := analysis.NewService(repo, nil, []int{1, 8})
	s.AddCalculatorByIndicator(stubCalculatorByIndicator{
		name:        "analytic",
		byIndicator: domain.EMAIndicator,
		depths:      map[int]bool{1: true},
		onCalculate: func(_ context.Context, indicator domain.Indicator, depth int) (*analysis.Analytic, error) {
			return &analysis.Analytic{
				ID:             uuid.New(),
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

	dt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	result, err := s.CalculateByIndicator(context.Background(), domain.Indicator{
		Exchange: "bybit",
		Symbol:   "BTCUSDT",
		Unit:     domain.HourUnit,
		Interval: 1,
		Datetime: dt,
		Name:     domain.EMAIndicator,
		Depth:    26,
		Value:    100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Value != 101 {
		t.Fatalf("unexpected result: %#v", result)
	}

	stored, err := repo.Find(context.Background(), "analytic", "bybit", "BTCUSDT", domain.HourUnit, dt, 1, 26, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stored == nil || stored.Value != 101 {
		t.Fatalf("expected stored value 101, got %#v", stored)
	}
}

func TestService_CalculateByIndicator_returnsCachedValue(t *testing.T) {
	repo := memory.NewAnalysisRepository(100)
	dt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	cached := analysis.Analytic{
		ID:             uuid.New(),
		Name:           "analytic",
		Exchange:       "bybit",
		Symbol:         "BTCUSDT",
		Unit:           domain.HourUnit,
		Interval:       1,
		Datetime:       dt,
		Depth:          1,
		ByIndicator:    domain.EMAIndicator,
		IndicatorDepth: 26,
		Value:          777,
	}
	if err := repo.Save(context.Background(), cached); err != nil {
		t.Fatalf("save fixture: %v", err)
	}

	var calls atomic.Int32
	s := analysis.NewService(repo, nil, []int{1})
	s.AddCalculatorByIndicator(stubCalculatorByIndicator{
		name:        "analytic",
		byIndicator: domain.EMAIndicator,
		depths:      map[int]bool{1: true},
		onCalculate: func(_ context.Context, _ domain.Indicator, _ int) (*analysis.Analytic, error) {
			calls.Add(1)
			return nil, nil
		},
	})

	result, err := s.CalculateByIndicator(context.Background(), domain.Indicator{
		Exchange: "bybit",
		Symbol:   "BTCUSDT",
		Unit:     domain.HourUnit,
		Interval: 1,
		Datetime: dt,
		Name:     domain.EMAIndicator,
		Depth:    26,
		Value:    100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Value != 777 {
		t.Fatalf("expected cached value 777, got %#v", result)
	}
	if calls.Load() != 0 {
		t.Fatalf("calculator must not be called, calls=%d", calls.Load())
	}
}

func TestService_AnalyticByIndicators_batchProcessing(t *testing.T) {
	repo := memory.NewAnalysisRepository(100)
	s := analysis.NewService(repo, nil, []int{1})
	s.AddCalculatorByIndicator(stubCalculatorByIndicator{
		name:        "analytic",
		byIndicator: domain.EMAIndicator,
		depths:      map[int]bool{1: true},
		onCalculate: func(_ context.Context, indicator domain.Indicator, depth int) (*analysis.Analytic, error) {
			return &analysis.Analytic{
				ID:             uuid.New(),
				Name:           "analytic",
				Exchange:       indicator.Exchange,
				Symbol:         indicator.Symbol,
				Unit:           indicator.Unit,
				Interval:       indicator.Interval,
				Datetime:       indicator.Datetime,
				Depth:          depth,
				ByIndicator:    indicator.Name,
				IndicatorDepth: indicator.Depth,
				Value:          float64(indicator.Datetime.Hour()),
			}, nil
		},
	})

	times := []time.Time{
		time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
	}
	indicators := make([]domain.Indicator, len(times))
	for i, dt := range times {
		indicators[i] = domain.Indicator{
			Exchange: "bybit",
			Symbol:   "BTCUSDT",
			Unit:     domain.HourUnit,
			Interval: 1,
			Datetime: dt,
			Name:     domain.EMAIndicator,
			Depth:    26,
			Value:    100,
		}
	}

	result, err := s.AnalyticByIndicators(context.Background(), indicators, "analytic", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 || result[0].Value != 8 || result[1].Value != 9 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func newServiceWithAnalyticCalculator(t *testing.T, calc stubCalculatorByAnalytic) *analysis.Service {
	t.Helper()
	s := analysis.NewService(memory.NewAnalysisRepository(100), nil, []int{1, 8, 9, 10})
	s.AddCalculatorByAnalytic(calc)
	return s
}

type stubCalculatorByAnalytic struct {
	name        string
	byAnalytic  string
	depths      map[int]bool
	onCalculate func(ctx context.Context, data analysis.Analytic, depth int) (*analysis.Analytic, error)
}

func (s stubCalculatorByAnalytic) Name() string             { return s.name }
func (s stubCalculatorByAnalytic) ByAnalytic() string       { return s.byAnalytic }
func (s stubCalculatorByAnalytic) SupportInterval(int) bool { return true }
func (s stubCalculatorByAnalytic) SupportDepth(depth int) bool {
	return s.depths[depth]
}
func (s stubCalculatorByAnalytic) Calculate(ctx context.Context, data analysis.Analytic, depth int) (*analysis.Analytic, error) {
	if s.onCalculate == nil {
		return nil, nil
	}
	return s.onCalculate(ctx, data, depth)
}

type stubCalculatorByIndicator struct {
	name        string
	byIndicator string
	depths      map[int]bool
	onCalculate func(ctx context.Context, indicator domain.Indicator, depth int) (*analysis.Analytic, error)
}

func (s stubCalculatorByIndicator) Name() string             { return s.name }
func (s stubCalculatorByIndicator) ByIndicator() string      { return s.byIndicator }
func (s stubCalculatorByIndicator) SupportInterval(int) bool { return true }
func (s stubCalculatorByIndicator) SupportDepth(depth int) bool {
	return s.depths[depth]
}
func (s stubCalculatorByIndicator) Calculate(ctx context.Context, indicator domain.Indicator, depth int) (*analysis.Analytic, error) {
	if s.onCalculate == nil {
		return nil, nil
	}
	return s.onCalculate(ctx, indicator, depth)
}

func sampleAnalytic(name string, dt time.Time, depth int) analysis.Analytic {
	return sampleAnalyticWithValue(name, dt, depth, 1)
}

func sampleAnalyticWithValue(name string, dt time.Time, depth int, value float64) analysis.Analytic {
	return analysis.Analytic{
		ID:             uuid.New(),
		Exchange:       "bybit",
		Symbol:         "BTCUSDT",
		Unit:           domain.HourUnit,
		Interval:       1,
		Name:           name,
		Datetime:       dt,
		Depth:          depth,
		ByIndicator:    domain.EMAIndicator,
		IndicatorDepth: 26,
		Value:          value,
	}
}

func derivedAnalytic(source analysis.Analytic, depth int, value float64) *analysis.Analytic {
	return &analysis.Analytic{
		ID:             uuid.New(),
		Name:           "derived",
		Exchange:       source.Exchange,
		Symbol:         source.Symbol,
		Unit:           source.Unit,
		Interval:       source.Interval,
		Datetime:       source.Datetime,
		Depth:          depth,
		ByIndicator:    source.Name,
		IndicatorDepth: source.Depth,
		Value:          value,
	}
}
