//go:build smoke

// Пакет smoke_test — минимальные проверки «живости» сервера.
//
// Зачем: быстрая валидация после деплоя (< 30 сек). Если smoke не проходит —
// сервис сломан и дальнейшие тесты бессмысленны.
//
// Что проверяют:
// - Сервер отвечает (TCP connection + HTTP 200)
// - Основные endpoint'ы доступны и возвращают ожидаемую структуру
// - Метрики отдаются (Prometheus endpoint)
//
// Запуск: make test-smoke (требует запущенный сервер)
package smoke_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func baseURL() string {
	if v := os.Getenv("TEST_API_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "http://localhost:8080"
}

func TestSmoke_ServerResponds(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL() + "/api/server")
	if err != nil {
		t.Fatalf("server not responding: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestSmoke_TradingEndpointsAvailable(t *testing.T) {
	endpoints := []struct {
		path    string
		payload string
	}{
		{
			"/api/v1/calculator/trading/avg-entry-price",
			`{"entry_volume":1,"entry_price":100,"new_volume":1,"new_price":200}`,
		},
		{
			"/api/v1/calculator/trading/spot-pnl",
			`{"volume":1,"entry_price":100,"mark_price":110}`,
		},
		{
			"/api/v1/calculator/trading/liquidation-price",
			`{"side":"long","volume":1,"entry_price":100000,"margin":10000}`,
		},
	}

	client := &http.Client{Timeout: 5 * time.Second}
	for _, ep := range endpoints {
		t.Run(ep.path, func(t *testing.T) {
			resp, err := client.Post(
				fmt.Sprintf("%s%s", baseURL(), ep.path),
				"application/json",
				strings.NewReader(ep.payload),
			)
			if err != nil {
				t.Fatalf("request to %s failed: %v", ep.path, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("%s: expected 200, got %d", ep.path, resp.StatusCode)
			}
		})
	}
}

func TestSmoke_ResponseFormat(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL() + "/api/server")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Fatalf("expected JSON content-type, got: %s", contentType)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
}

func TestSmoke_MetricsEndpoint(t *testing.T) {
	metricsURL := os.Getenv("TEST_METRICS_URL")
	if metricsURL == "" {
		metricsURL = "http://localhost:8080/metrics"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(metricsURL)
	if err != nil {
		t.Skipf("metrics endpoint not available: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("metrics: expected 200, got %d", resp.StatusCode)
	}
}
