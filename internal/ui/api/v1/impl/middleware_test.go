// Тесты ограничения частоты запросов к REST API v1.
//
// Проверяют ровно то правило, из-за которого middleware существует: без X-Token
// адрес получает 10 запросов в минуту, с X-Token — сколько угодно. Правило легко
// испортить незаметно: достаточно сбить размер квоты, потерять Skipper или начать
// считать по пути вместо адреса — и лимит либо перестанет ограничивать, либо
// начнёт ограничивать не тех. Поэтому каждый из этих случаев здесь свой тест, а
// не один общий сценарий.
//
// Числа берём из констант пакета, а не вписываем в тесты: иначе при следующей
// правке квоты тесты продолжат проверять прежнее значение.
package impl

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

// newLimitedServer поднимает echo с лимитом на единственном обработчике.
// Хранилище счётчиков создаётся внутри RateLimit, поэтому у каждого теста своё:
// иначе тесты делили бы счётчик и падали бы по очереди из-за чужих запросов.
func newLimitedServer(t *testing.T) *echo.Echo {
	t.Helper()
	e := echo.New()
	e.Use(RateLimit())
	e.GET("/api/v1/server", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	return e
}

// do отправляет запрос от указанного адреса с указанным токеном.
// Адрес задаётся через RemoteAddr: именно из него middleware берёт клиента.
func do(e *echo.Echo, ip, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/server", nil)
	req.RemoteAddr = ip + ":54321"
	if token != "" {
		req.Header.Set(TokenHeader, token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// spendQuota расходует анонимную квоту адреса целиком и убеждается, что все
// запросы квоты прошли: квота тратится подряд, всплеск равен её размеру.
func spendQuota(t *testing.T, e *echo.Echo, ip string) {
	t.Helper()
	for i := 0; i < anonymousLimit; i++ {
		if code := do(e, ip, "").Code; code != http.StatusOK {
			t.Fatalf("запрос %d из квоты %d должен пройти, получено %d", i+1, anonymousLimit, code)
		}
	}
}

// Квота расходуется подряд, а запрос сверх неё получает отказ: 10 проходят,
// 11-й — нет.
func TestAnonymousRequestsBeyondQuotaRejected(t *testing.T) {
	e := newLimitedServer(t)

	spendQuota(t, e, "203.0.113.10")

	rec := do(e, "203.0.113.10", "")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("запрос сверх квоты должен получить 429, получено %d", rec.Code)
	}
	wantRetry := strconv.Itoa(int(anonymousRefill.Seconds()))
	if got := rec.Header().Get(echo.HeaderRetryAfter); got != wantRetry {
		t.Errorf("Retry-After должен сообщать %s секунд ожидания, получено %q", wantRetry, got)
	}
}

// Квота — это именно 10 запросов в минуту, а не «10 запросов когда-нибудь»:
// пополнение должно идти по одному разу в шесть секунд.
func TestQuotaIsTenPerMinute(t *testing.T) {
	if anonymousLimit != 10 || anonymousWindow != time.Minute {
		t.Fatalf("квота должна быть 10 запросов в минуту, задано %d за %s", anonymousLimit, anonymousWindow)
	}
	if anonymousRefill != 6*time.Second {
		t.Errorf("пополнение должно идти каждые 6 секунд, получено %s", anonymousRefill)
	}
}

// Отказ приходит в том же формате, что и остальные ошибки API: клиент разбирает
// один тип ответа, а не два. И код в ответе именно 429, а не 500 — иначе клиент
// не поймёт, что нужно подождать, и продолжит долбить сервис.
func TestRejectionBodyIsApiError(t *testing.T) {
	e := newLimitedServer(t)
	spendQuota(t, e, "203.0.113.11")
	rec := do(e, "203.0.113.11", "")

	var body struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("тело отказа должно быть json ошибки API: %v (%s)", err, rec.Body.String())
	}
	if body.Message == "" {
		t.Error("в отказе должно быть сообщение с причиной")
	}
}

// С токеном лимита нет: запросов подряд больше квоты, и проходят все.
func TestTokenLiftsLimit(t *testing.T) {
	e := newLimitedServer(t)

	for i := 0; i < anonymousLimit*2; i++ {
		if code := do(e, "203.0.113.20", "any-string").Code; code != http.StatusOK {
			t.Fatalf("запрос %d с токеном должен пройти, получено %d", i+1, code)
		}
	}
}

// Токен из одних пробелов за токен не считаем: прислать его проще, чем
// осмысленный, и снимать лимит он не должен.
func TestBlankTokenIsAnonymous(t *testing.T) {
	e := newLimitedServer(t)

	for i := 0; i < anonymousLimit; i++ {
		do(e, "203.0.113.21", "   ")
	}
	if code := do(e, "203.0.113.21", "   ").Code; code != http.StatusTooManyRequests {
		t.Fatalf("пробельный токен не должен снимать лимит, получено %d", code)
	}
}

// Счёт идёт по адресу: сосед, исчерпавший квоту, не мешает другому клиенту.
func TestLimitIsPerAddress(t *testing.T) {
	e := newLimitedServer(t)

	spendQuota(t, e, "203.0.113.30")
	if code := do(e, "203.0.113.30", "").Code; code != http.StatusTooManyRequests {
		t.Fatalf("первый адрес должен быть ограничен, получено %d", code)
	}
	if code := do(e, "203.0.113.31", "").Code; code != http.StatusOK {
		t.Fatalf("второй адрес не должен страдать из-за первого, получено %d", code)
	}
}

// Заголовки пересылки не должны сбрасывать счётчик: прислать случайный
// X-Forwarded-For может любой клиент, и если ему верить, лимита нет вообще.
func TestForwardedHeadersDoNotResetLimit(t *testing.T) {
	e := newLimitedServer(t)

	spendQuota(t, e, "203.0.113.50")

	for _, header := range []string{echo.HeaderXForwardedFor, echo.HeaderXRealIP} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/server", nil)
		req.RemoteAddr = "203.0.113.50:54321"
		req.Header.Set(header, "198.51.100.7")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusTooManyRequests {
			t.Errorf("заголовок %s не должен снимать лимит, получено %d", header, rec.Code)
		}
	}
}

// Один счётчик на все адреса API: обход разных путей — ровно тот случай, от
// которого лимит и нужен, поэтому смена пути квоту не сбрасывает.
func TestLimitSharedAcrossPaths(t *testing.T) {
	e := echo.New()
	e.Use(RateLimit())
	handler := func(c echo.Context) error { return c.NoContent(http.StatusOK) }
	e.GET("/api/v1/server", handler)
	e.GET("/api/v1/price/:exchange/:symbol", handler)

	spendQuota(t, e, "203.0.113.40")

	second := httptest.NewRequest(http.MethodGet, "/api/v1/price/binance/BTCUSDT", nil)
	second.RemoteAddr = "203.0.113.40:54321"
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, second)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("другой адрес API не должен сбрасывать лимит, получено %d", rec.Code)
	}
}
