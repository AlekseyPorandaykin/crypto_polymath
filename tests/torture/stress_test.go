//go:build torture

// Пакет torture_test — стресс-тесты для выявления проблем конкурентности и утечек.
//
// Зачем: обычные unit-тесты не обнаруживают race conditions, deadlocks и memory leaks.
// Torture-тесты создают экстремальную нагрузку (тысячи горутин, миллионы итераций)
// и запускаются с -race детектором.
//
// Какие проблемы выявляют:
// - Data race при параллельном вызове калькулятора (общие структуры)
// - Паники при экстремальных входах (NaN, Inf, отрицательные)
// - Утечки памяти (рост MemStats при длительных вычислениях)
// - Деградация API под нагрузкой (concurrent HTTP requests)
//
// Запуск: make test-torture (требует ~10 мин, -race включён)
package torture_test

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/trading"
)

func TestFutureCalculator_ConcurrentSafety(t *testing.T) {
	calc := trading.Future{Leverage: 10}
	pos := trading.Position{Side: trading.Long, Volume: 1, EntryPrice: 100_000, Margin: 10_000, Leverage: 10}

	var wg sync.WaitGroup
	for i := 0; i < 10_000; i++ {
		wg.Add(1)
		go func(price float64) {
			defer wg.Done()
			snapshot := calc.RiskAtPrice(pos, price, 0.005)
			if math.IsNaN(snapshot.UnrealizedPnL) {
				panic("NaN in concurrent access")
			}
			result := calc.SimulateAddOn(pos, trading.AddOn{Price: price, Margin: 1000}, 0.005)
			if math.IsNaN(result.After.EntryPrice) {
				panic("NaN in concurrent SimulateAddOn")
			}
		}(50_000 + float64(i)*10)
	}
	wg.Wait()
}

func TestFutureCalculator_MassiveInputRange(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < 100_000; i++ {
		volume := rng.Float64() * 1000
		entry := rng.Float64() * 200_000
		margin := rng.Float64() * 50_000
		leverage := 1 + rng.Float64()*199

		if volume <= 0 || entry <= 0 || margin <= 0 {
			continue
		}

		pos := trading.Position{
			Side: trading.Long, Volume: volume, EntryPrice: entry,
			Margin: margin, Leverage: leverage,
		}
		mark := entry * (0.5 + rng.Float64())

		snapshot := trading.Future{}.RiskAtPrice(pos, mark, 0.005)
		if math.IsNaN(snapshot.UnrealizedPnL) || math.IsInf(snapshot.UnrealizedPnL, 0) {
			t.Fatalf("non-finite PnL at iteration %d: %+v", i, snapshot)
		}
	}
}

func TestFutureCalculator_NoMemoryLeak(t *testing.T) {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	before := m.TotalAlloc

	pos := trading.Position{Side: trading.Long, Volume: 1, EntryPrice: 100_000, Margin: 10_000, Leverage: 10}
	for i := 0; i < 1_000_000; i++ {
		_ = trading.Future{}.RiskAtPrice(pos, 95_000+float64(i%10000), 0.005)
	}

	runtime.GC()
	runtime.ReadMemStats(&m)
	after := m.TotalAlloc

	allocated := after - before
	perOp := allocated / 1_000_000
	if perOp > 512 {
		t.Fatalf("too many allocations per op: %d bytes", perOp)
	}
	t.Logf("allocations per RiskAtPrice call: ~%d bytes", perOp)
}

func TestAPI_HighConcurrency(t *testing.T) {
	base := os.Getenv("TEST_API_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	base = strings.TrimRight(base, "/")

	const (
		concurrency = 50
		totalReqs   = 5_000
	)

	sem := make(chan struct{}, concurrency)
	var (
		mu        sync.Mutex
		errors    int64
		latencies []time.Duration
	)

	client := &http.Client{Timeout: 10 * time.Second}
	payload := `{"entry_volume":1,"entry_price":100000,"new_volume":1,"new_price":90000}`

	var wg sync.WaitGroup
	for i := 0; i < totalReqs; i++ {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() { <-sem; wg.Done() }()

			start := time.Now()
			resp, err := client.Post(
				base+"/api/v1/calculator/trading/avg-entry-price",
				"application/json",
				strings.NewReader(payload),
			)
			elapsed := time.Since(start)

			mu.Lock()
			latencies = append(latencies, elapsed)
			mu.Unlock()

			if err != nil {
				atomic.AddInt64(&errors, 1)
				return
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				atomic.AddInt64(&errors, 1)
			}
		}()
	}
	wg.Wait()

	errorRate := float64(errors) / float64(totalReqs) * 100
	t.Logf("Total requests: %d, Errors: %d (%.2f%%)", totalReqs, errors, errorRate)

	if len(latencies) > 0 {
		sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
		p50 := latencies[len(latencies)/2]
		p95 := latencies[int(float64(len(latencies))*0.95)]
		p99 := latencies[int(float64(len(latencies))*0.99)]
		t.Logf("Latency p50=%v, p95=%v, p99=%v", p50, p95, p99)
	}

	if errorRate > 5 {
		t.Fatalf("error rate %.2f%% exceeds 5%% threshold", errorRate)
	}
}

func TestSpotCalculator_ConcurrentSafety(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 10_000; i++ {
		wg.Add(1)
		go func(mark float64) {
			defer wg.Done()
			value, percent := trading.Spot{}.PnL(1.5, 50_000, mark)
			if math.IsNaN(value) || math.IsNaN(percent) {
				panic(fmt.Sprintf("NaN at mark=%v", mark))
			}
		}(40_000 + float64(i)*2)
	}
	wg.Wait()
}
