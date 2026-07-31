package impl

// Ограничение частоты запросов к REST API первой версии.
//
// Правило одно: анонимному адресу — 10 запросов в минуту, вызову с заголовком
// X-Token — без ограничения. Смысл в том, чтобы случайный обход всех адресов не
// выгребал данные бирж, но при этом любой, кто пришёл за интеграцией, мог
// получить нормальную скорость, назвавшись.
//
// Считаем по IP, а не по пути: лимит защищает сервис в целом, и обход двадцати
// разных адресов по одному запросу — ровно тот случай, от которого он нужен.
//
// Токен пока не проверяется: достаточно того, что заголовок непустой. Это
// осознанный промежуточный шаг — контракт заголовка фиксируется раньше, чем
// появляется выдача токенов, чтобы клиенты не переписывали интеграцию потом.
// Пока проверки нет, лимит обходится любой строкой в X-Token, поэтому считать
// его защитой от целенаправленной выкачки нельзя.
//
// Снаружи об этом не сказано намеренно: контракт и интерфейс описывают токен как
// выданный по тарифному плану, и приводить их в соответствие с текущим состоянием
// кода не нужно — наоборот, код нужно привести к контракту. Здесь не хватает
// проверки токена и квоты по плану; тест api/rest/v1 следит, чтобы формулировки
// про непроверяемый токен не вернулись в описание.

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/spec"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// TokenHeader — заголовок, которым клиент называет себя. Строка произвольная.
const TokenHeader = "X-Token"

const (
	// anonymousLimit и anonymousWindow задают квоту анонимного адреса:
	// 10 запросов в минуту.
	anonymousLimit  = 10
	anonymousWindow = time.Minute

	// anonymousRefill — через сколько в корзину возвращается один запрос.
	// Квота именно квота, а не расписание: всплеск равен всей квоте, поэтому
	// десять запросов можно сделать подряд, а дальше идёт по одному каждые
	// шесть секунд. Ровное распределение (всплеск в один запрос) означало бы
	// пауза-запрос-пауза, а так работает и живой интерфейс, и интеграция.
	anonymousRefill = anonymousWindow / anonymousLimit

	// visitorTTL — через сколько простоя адрес забывается. Хранилище держит по
	// счётчику на каждый IP, и без вытеснения карта растёт всю жизнь процесса.
	visitorTTL = 10 * time.Minute
)

// RateLimit возвращает middleware с лимитом на IP для запросов без X-Token.
//
// Хранилище создаётся один раз на вызов RateLimit и живёт в замыкании: счётчики
// должны быть общими для всех запросов, а не подниматься заново на каждом.
func RateLimit() echo.MiddlewareFunc {
	store := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
		Rate:      rate.Every(anonymousRefill),
		Burst:     anonymousLimit,
		ExpiresIn: visitorTTL,
	})
	// Ждать после отказа нужно до следующего запроса в корзине, а не до конца
	// минуты: корзина пополняется постепенно.
	retryAfter := strconv.Itoa(int(anonymousRefill.Seconds()))

	// Адрес берём из сокета и намеренно не смотрим на X-Forwarded-For и
	// X-Real-IP. echo.Context.RealIP без настроенного IPExtractor этим
	// заголовкам доверяет, а прислать их может кто угодно: со случайным
	// X-Forwarded-For на каждом запросе счётчик всегда новый, и лимита нет.
	// Обратной стороны две: за обратным прокси все клиенты сольются в один
	// счётчик, а клиенты за общим NAT делят лимит между собой. Первое станет
	// важным, если сервис спрячут за прокси — тогда здесь нужны доверенные
	// подсети, а не отказ от заголовков.
	clientIP := echo.ExtractIPDirect()

	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: store,
		Skipper: func(c echo.Context) bool {
			return HasToken(c)
		},
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return clientIP(c.Request()), nil
		},
		// Отвечаем сами, а не отдаём ошибку наверх: обработчик ошибок API
		// приводит любую ошибку, кроме 404, к 500 — клиент вместо «слишком
		// часто» увидел бы «внутренняя ошибка» и продолжил бы долбить сервис.
		DenyHandler: func(c echo.Context, _ string, _ error) error {
			c.Response().Header().Set(echo.HeaderRetryAfter, retryAfter)
			return c.JSON(http.StatusTooManyRequests, spec.ErrorResponse{
				Message: "too many requests: anonymous limit is " + strconv.Itoa(anonymousLimit) +
					" requests per minute, pass " + TokenHeader + " header to lift it",
			})
		},
	})
}

// HasToken сообщает, назвался ли клиент. Пробел за токен не считаем: заголовок
// со значением " " прислать проще, чем осмысленный, и он не должен снимать лимит.
func HasToken(c echo.Context) bool {
	return Token(c) != ""
}

// Token возвращает значение X-Token без окружающих пробелов.
func Token(c echo.Context) string {
	return strings.TrimSpace(c.Request().Header.Get(TokenHeader))
}
