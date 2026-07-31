// Тесты приёма сообщений контактной формы.
//
// Доставки писем пока нет, поэтому проверять «дошло ли письмо» нечего — но именно
// поэтому проверки входных данных здесь важнее обычного: единственное, что делает
// обработчик, это отсеивает мусор и сохраняет остальное в лог. Если проверка
// разойдётся с контрактом, посетитель увидит либо отказ там, где форма разрешает
// ввод, либо наоборот — принятое сообщение, которое потом нельзя прочитать.
//
// Отдельный тест закрывает код ответа: 202, а не 200. Обещание «приняли, ответим»
// отличается от «обработали», и клиенты формы на этом различии построены.
package impl

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/spec"
	"github.com/labstack/echo/v4"
)

// postContact отправляет тело в обработчик формы напрямую, без группы и лимита:
// лимит проверяется отдельно в middleware_test.go.
func postContact(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/contact", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	if err := (&Handler{}).PostContact(e.NewContext(req, rec)); err != nil {
		t.Fatalf("PostContact вернул ошибку: %v", err)
	}
	return rec
}

func TestPostContactAcceptsMessage(t *testing.T) {
	rec := postContact(t, `{
		"email": "trader@example.com",
		"subject": "Historical candlesticks",
		"message": "Is a longer history available through the API?"
	}`)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("принятое сообщение должно давать %d, получено %d", http.StatusAccepted, rec.Code)
	}
	var resp spec.ContactMessageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("разбор ответа: %v", err)
	}
	if resp.Message == "" {
		t.Error("в ответе нет подтверждения: посетителю нечего показать в форме")
	}
}

// Каждая причина отказа названа в ответе: форма показывает текст сервера как есть,
// поэтому пустое или общее сообщение оставило бы посетителя без подсказки.
func TestPostContactRejectsInvalidMessage(t *testing.T) {
	cases := map[string]string{
		"адрес без домена":       `{"email": "trader", "subject": "Question", "message": "Tell me more about the API"}`,
		"адрес с именем":         `{"email": "Ann <ann@example.com>", "subject": "Question", "message": "Tell me more about the API"}`,
		"пустой адрес":           `{"email": "", "subject": "Question", "message": "Tell me more about the API"}`,
		"тема из пробелов":       `{"email": "a@example.com", "subject": "   ", "message": "Tell me more about the API"}`,
		"слишком короткая тема":  `{"email": "a@example.com", "subject": "Hi", "message": "Tell me more about the API"}`,
		"слишком короткий текст": `{"email": "a@example.com", "subject": "Question", "message": "short"}`,
		"тело не json":           `not json at all`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			rec := postContact(t, body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("ожидался %d, получено %d", http.StatusBadRequest, rec.Code)
			}
			var resp spec.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("разбор ответа: %v", err)
			}
			if resp.Message == "" {
				t.Error("отказ без объяснения: форме нечего показать посетителю")
			}
		})
	}
}

// Границы длин совпадают с контрактом и с атрибутами maxlength в форме. Тест
// проверяет сами края: ровно на границе сообщение принимается, за ней — нет.
func TestPostContactLengthBoundaries(t *testing.T) {
	message := strings.Repeat("a", contactMessageMaxLen)
	subject := strings.Repeat("s", contactSubjectMaxLen)

	rec := postContact(t, `{"email": "a@example.com", "subject": "`+subject+`", "message": "`+message+`"}`)
	if rec.Code != http.StatusAccepted {
		t.Errorf("сообщение ровно по границе должно приниматься, получено %d", rec.Code)
	}

	rec = postContact(t, `{"email": "a@example.com", "subject": "`+subject+`", "message": "`+message+`a"}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("текст длиннее границы должен отклоняться, получено %d", rec.Code)
	}

	rec = postContact(t, `{"email": "a@example.com", "subject": "`+subject+`s", "message": "`+message+`"}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("тема длиннее границы должна отклоняться, получено %d", rec.Code)
	}
}

// Форма — такой же публичный метод, как остальные, и лимит на неё распространяется
// сам собой: он висит на группе /api/v1 целиком. Тест закрепляет это свойство,
// потому что теряется оно незаметно — достаточно зарегистрировать путь мимо группы.
func TestPostContactIsRateLimited(t *testing.T) {
	e := echo.New()
	group := e.Group("/api/v1")
	group.Use(RateLimit())
	spec.RegisterHandlers(group, &Handler{})

	body := `{"email":"a@example.com","subject":"Question","message":"Tell me more about the API"}`
	send := func() int {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/contact", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.RemoteAddr = "203.0.113.7:54321"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec.Code
	}

	for i := 0; i < anonymousLimit; i++ {
		if code := send(); code != http.StatusAccepted {
			t.Fatalf("сообщение %d из квоты должно приниматься, получено %d", i+1, code)
		}
	}
	if code := send(); code != http.StatusTooManyRequests {
		t.Errorf("сообщение за пределами квоты должно давать %d, получено %d", http.StatusTooManyRequests, code)
	}
}

// Пробелы по краям — обычное следствие копирования текста в поле, и отказывать
// из-за них нельзя: обработчик обрезает их сам, а не считает частью значения.
func TestPostContactTrimsFields(t *testing.T) {
	email, subject, message, err := validateContactMessage(spec.ContactMessageRequest{
		Email:   "  trader@example.com  ",
		Subject: "  Question  ",
		Message: "  Tell me more about the API  ",
	})
	if err != nil {
		t.Fatalf("сообщение с пробелами по краям должно приниматься: %v", err)
	}
	if email != "trader@example.com" || subject != "Question" || message != "Tell me more about the API" {
		t.Errorf("поля не обрезаны: %q, %q, %q", email, subject, message)
	}
}
