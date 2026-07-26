//go:build acceptance

// Пакет acceptance_test — приёмочные (end-to-end) тесты HTTP API.
//
// Зачем: проверяют, что API работает корректно как единое целое —
// от HTTP-запроса до ответа. Запускаются на реальном (или docker) сервере.
//
// Какие проблемы выявляют:
// - Неработающие endpoint'ы после рефакторинга
// - Некорректный формат ответа (структура JSON, статус-коды)
// - Проблемы сериализации/десериализации (NaN, пустые поля)
// - Регрессии в бизнес-логике на уровне интеграции
//
// Запуск: make test-acceptance (требует запущенный docker-compose)
// Переменная окружения: TEST_API_URL (по умолчанию http://localhost:8080)
package acceptance_test

import (
	"encoding/json"
	"fmt"
	"io"
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

func TestAPI_ServerInfo(t *testing.T) {
	resp, err := httpGet(t, "/api/server")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := body["exchanges"]; !ok {
		t.Fatal("response missing 'exchanges' field")
	}
	if _, ok := body["symbols"]; !ok {
		t.Fatal("response missing 'symbols' field")
	}
}

func TestAPI_TradingAvgEntryPrice(t *testing.T) {
	payload := `{"entry_volume":1,"entry_price":100,"new_volume":1,"new_price":200}`
	resp, err := httpPost(t, "/api/v1/calculator/trading/avg-entry-price", payload)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	avgPrice, ok := body["avg_entry_price"].(float64)
	if !ok {
		t.Fatalf("missing avg_entry_price in response: %v", body)
	}
	if avgPrice < 149 || avgPrice > 151 {
		t.Fatalf("expected ~150, got %v", avgPrice)
	}
}

func TestAPI_TradingLiquidationPrice(t *testing.T) {
	payload := `{"side":"long","volume":1,"entry_price":100000,"margin":10000,"maintenance_margin_rate":0.005}`
	resp, err := httpPost(t, "/api/v1/calculator/trading/liquidation-price", payload)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := body["liquidation_price"]; !ok {
		t.Fatalf("missing liquidation_price: %v", body)
	}
}

func TestAPI_TradingSimulateAddOn(t *testing.T) {
	payload := `{
		"position":{"side":"long","volume":1,"entry_price":100000,"margin":10000,"leverage":10},
		"add_on":{"price":90000,"margin":10000}
	}`
	resp, err := httpPost(t, "/api/v1/calculator/trading/simulate-add-on", payload)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := body["volume_added"]; !ok {
		t.Fatalf("missing volume_added: %v", body)
	}
}

func TestAPI_TradingRiskAtPrice(t *testing.T) {
	payload := `{
		"position":{"side":"long","volume":1,"entry_price":100000,"margin":10000,"leverage":10},
		"mark_price":95000
	}`
	resp, err := httpPost(t, "/api/v1/calculator/trading/risk-at-price", payload)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := body["unrealized_pnl"]; !ok {
		t.Fatalf("missing unrealized_pnl: %v", body)
	}
}

func TestAPI_TradingSpotPnL(t *testing.T) {
	payload := `{"volume":2,"entry_price":100,"mark_price":150}`
	resp, err := httpPost(t, "/api/v1/calculator/trading/spot-pnl", payload)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	val, ok := body["value"].(float64)
	if !ok {
		t.Fatalf("missing value: %v", body)
	}
	if val < 99 || val > 101 {
		t.Fatalf("expected ~100, got %v", val)
	}
}

func TestAPI_TradingBadRequest(t *testing.T) {
	payload := `{"invalid": true}`
	resp, err := httpPost(t, "/api/v1/calculator/trading/avg-entry-price", payload)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var result map[string]any
		_ = json.Unmarshal(body, &result)
		if avg, ok := result["avg_entry_price"].(float64); ok && avg != 0 {
			t.Fatalf("expected zero or error for invalid request, got %v", avg)
		}
	}
}

func TestAPI_NotFoundEndpoint(t *testing.T) {
	resp, err := httpGet(t, "/api/v1/nonexistent")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 404/405, got %d", resp.StatusCode)
	}
}

func httpGet(t *testing.T, path string) (*http.Response, error) {
	t.Helper()
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s%s", baseURL(), path)
	return client.Get(url)
}

func httpPost(t *testing.T, path, body string) (*http.Response, error) {
	t.Helper()
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s%s", baseURL(), path)
	return client.Post(url, "application/json", strings.NewReader(body))
}

func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d; body: %s", expected, resp.StatusCode, string(body))
	}
}
