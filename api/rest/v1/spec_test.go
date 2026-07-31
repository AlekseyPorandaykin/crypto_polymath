// Тесты контракта REST API первой версии.
//
// Контракт — единственное, что видит внешний клиент: по нему генерируют клиентов
// и по нему решают, как обращаться к сервису. Поведение, о котором в контракте не
// написано, для клиента не существует, поэтому ограничение частоты запросов
// должно быть описано, а не только реализовано в обработчике.
//
// Тесты закрывают расхождение контракта с поведением сервера:
//
//   - заголовок X-Token объявлен и объявлен необязательным: иначе генераторы
//     клиентов потребуют токен, которого у клиента нет;
//   - ответ 429 описан у каждой операции, а не у части: лимит считается по адресу
//     и одинаково срабатывает на любом методе;
//   - числа лимита названы в обзоре — клиенту нужно знать, сколько ждать.
package v1

import (
	"regexp"
	"strings"
	"testing"
)

// Операции лежат на четырёх пробелах: сами пути — на двух, метод внутри пути — на
// четырёх. Считаем по отступу, а не по слову: get и post встречаются и в тексте
// описаний.
var operationRe = regexp.MustCompile(`(?m)^ {4}(get|post):$`)

// Ответы операции лежат на восьми пробелах, как и остальные коды рядом.
var tooManyRequestsRe = regexp.MustCompile(`(?m)^ {8}429:$`)

// Заголовок должен быть объявлен схемой апи-ключа: это единственный способ
// показать его в документации у каждой операции, не превращая в обязательный
// параметр каждого запроса.
func TestContractDeclaresTokenHeader(t *testing.T) {
	spec := string(Specification)

	if !strings.Contains(spec, "securitySchemes:") {
		t.Fatal("в контракте нет securitySchemes — заголовок X-Token негде объявить")
	}
	for _, want := range []string{"TokenHeader:", "type: apiKey", "in: header", "name: X-Token"} {
		if !strings.Contains(spec, want) {
			t.Errorf("схема заголовка описана не полностью: нет %q", want)
		}
	}
}

// Пустое требование в списке security — то, чем в OpenAPI выражается
// «аутентификация необязательна». Без него клиент по контракту обязан прислать
// токен, а сервер его не требует: контракт разошёлся бы с поведением.
func TestContractKeepsAnonymousCallsAllowed(t *testing.T) {
	spec := string(Specification)

	block := regexp.MustCompile(`(?m)^security:\n(?: +- .*\n)+`).FindString(spec)
	if block == "" {
		t.Fatal("в контракте нет корневого блока security — токен объявлен, но не привязан к операциям")
	}
	if !regexp.MustCompile(`- \{ *\}`).MatchString(block) {
		t.Errorf("в security нет пустого требования: анонимный вызов оказался запрещён контрактом\n%s", block)
	}
	if !strings.Contains(block, "TokenHeader:") {
		t.Errorf("в security нет требования с токеном: заголовок не попадёт в документацию операций\n%s", block)
	}
}

// 429 должен быть описан у всех операций: лимит считается по адресу и не зависит
// от того, какой метод вызывают. Пропуск даже у одной операции — это клиент,
// который не ожидает отказа и не умеет ждать.
func TestEveryOperationDocumentsRateLimitResponse(t *testing.T) {
	spec := string(Specification)

	operations := len(operationRe.FindAllString(spec, -1))
	rejections := len(tooManyRequestsRe.FindAllString(spec, -1))

	if operations == 0 {
		t.Fatal("в контракте не нашлось ни одной операции — сломался разбор, а не контракт")
	}
	if rejections != operations {
		t.Errorf("ответ 429 описан у %d операций из %d", rejections, operations)
	}
}

// Обзор должен называть сами числа: клиент, получивший отказ, узнаёт из контракта
// сколько ждать и что делать, чтобы отказ не повторялся.
func TestContractExplainsLimitNumbers(t *testing.T) {
	spec := string(Specification)

	for _, want := range []string{"X-Token", "10 requests per minute", "Retry-After", "429"} {
		if !strings.Contains(spec, want) {
			t.Errorf("в описании контракта не сказано про %q", want)
		}
	}
}

// Лимит владельца токена зависит от тарифного плана, и контракт обещает не
// «без ограничений», а «по плану»: клиент, который построит интеграцию на
// отсутствии лимита, сломается в день включения тарифных квот.
func TestContractTiesTokenLimitToPlan(t *testing.T) {
	spec := string(Specification)

	if !strings.Contains(spec, "depends on the pricing plan") {
		t.Error("в контракте не сказано, что лимит с токеном зависит от тарифного плана")
	}
	if strings.Contains(spec, "| with `X-Token` | not limited |") {
		t.Error("в таблице лимитов у токена всё ещё обещано отсутствие ограничений")
	}
}

// Единственный метод, который принимает данные от человека, а не отдаёт рыночные:
// контактная форма лендинга. Проверяем то, на что опираются и форма, и обработчик:
// путь, обязательные поля, границы длин и код 202.
//
// Код ответа здесь часть смысла, а не деталь реализации: 202 говорит «приняли к
// отправке», и подтверждение в интерфейсе обещает именно это. Заменив его на 200,
// сервис начал бы утверждать, что сообщение обработано.
func TestContractDescribesContactEndpoint(t *testing.T) {
	spec := string(Specification)

	if !strings.Contains(spec, "\n  /contact:\n") {
		t.Fatal("в контракте нет пути /contact")
	}
	contact := regexp.MustCompile(`(?s)\n  /contact:\n.*?\n  /`).FindString(spec)
	if contact == "" {
		t.Fatal("не удалось выделить описание /contact")
	}
	if !strings.Contains(contact, "        202:") {
		t.Error("у /contact нет ответа 202: форма обещает приём, а не обработку")
	}
	if strings.Contains(contact, "        200:") {
		t.Error("у /contact объявлен 200: приём сообщения не равен его обработке")
	}
	for _, want := range []string{"ContactMessageRequest", "ContactMessageResponse", "429:", "400:"} {
		if !strings.Contains(contact, want) {
			t.Errorf("в описании /contact нет %q", want)
		}
	}

	schema := regexp.MustCompile(`(?s)ContactMessageRequest:\n.*?ContactMessageResponse:`).FindString(spec)
	if schema == "" {
		t.Fatal("в контракте нет схемы ContactMessageRequest")
	}
	if !strings.Contains(schema, "required: [ email, subject, message ]") {
		t.Error("в схеме сообщения не все три поля обязательны")
	}
	// Границы длин объявлены в контракте намеренно: по ним форма в браузере
	// проверяет ввод до запроса, а обработчик — после.
	for _, limit := range []string{"maxLength: 254", "minLength: 3", "maxLength: 150", "minLength: 10", "maxLength: 5000"} {
		if !strings.Contains(schema, limit) {
			t.Errorf("в схеме сообщения нет ограничения %q", limit)
		}
	}
}

// Контракт читают снаружи, и читают целиком: описания уходят в документацию на
// /docs/api и в сгенерированных клиентов, а YAML-комментарии видит всякий, кто
// откроет файл по /docs/api/openapi.yaml. Поэтому язык здесь один — английский,
// включая комментарии.
func TestContractIsEnglish(t *testing.T) {
	cyrillic := regexp.MustCompile(`\p{Cyrillic}+[^\n]*`)

	for _, found := range cyrillic.FindAllString(string(Specification), -1) {
		t.Errorf("в контракте непереведённый текст: %q", found)
	}
}

// Снаружи токен описан как выданный по тарифному плану, и текущее состояние
// проверки в контракт не просачивается: клиент должен получать токен, а не
// подставлять произвольную строку. Формулировки вроде «не проверяется» или
// «произвольная строка» легко вернуть при правке описания, поэтому их отсутствие
// закреплено тестом, а не только договорённостью.
func TestContractDoesNotAdvertiseUncheckedToken(t *testing.T) {
	spec := strings.ToLower(string(Specification))

	for _, forbidden := range []string{
		"not validated",
		"not enforced",
		"arbitrary",
		"non-empty",
	} {
		if strings.Contains(spec, forbidden) {
			t.Errorf("в контракте снова сказано %q — снаружи состояние проверки токена не раскрываем", forbidden)
		}
	}
}
