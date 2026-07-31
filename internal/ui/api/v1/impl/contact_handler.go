package impl

// Приём сообщений из контактной формы лендинга.
//
// Доставки пока нет: письма никуда не уходят, потому что у сервиса ещё нет ни
// почтового транспорта, ни адресата. Но принимать сообщения нужно уже сейчас —
// иначе кнопка на лендинге ведёт в пустоту.
//
// Поэтому сообщение проверяется и пишется в лог. Это не заглушка ради ответа
// 202: запись в лог означает, что письмо можно прочитать и ответить руками, то
// есть обещание «сообщение принято» выполняется, пусть и вручную. Ответ 202, а
// не 200, ровно об этом и говорит: приняли, но не утверждаем, что уже прочитали.
//
// Проверки повторяют ограничения контракта, а не заменяют их: сгенерированные
// типы длину и формат адреса не контролируют, а форма в браузере проверяет то же
// самое — от посетителя, а не от клиента API.

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/AlekseyPorandaykin/crypto_polymath/internal/ui/api/v1/spec"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Границы длин совпадают с minLength/maxLength в контракте и с maxlength в форме
// лендинга. Слишком короткая тема или текст — почти всегда случайное нажатие, а
// верхние границы держат размер записи в логе предсказуемым.
const (
	contactEmailMaxLen   = 254
	contactSubjectMinLen = 3
	contactSubjectMaxLen = 150
	contactMessageMinLen = 10
	contactMessageMaxLen = 5000
)

var (
	errContactEmailInvalid   = errors.New("email is not a valid address")
	errContactSubjectLength  = errors.New("subject length must be between 3 and 150 characters")
	errContactMessageLength  = errors.New("message length must be between 10 and 5000 characters")
	errContactMessageInvalid = errors.New("invalid request body")
)

func (h *Handler) PostContact(ctx echo.Context) error {
	var req spec.ContactMessageRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: errContactMessageInvalid.Error()})
	}
	email, subject, message, err := validateContactMessage(req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, spec.ErrorResponse{Message: err.Error()})
	}

	// Пока нет доставки, лог — единственное место, где сообщение сохраняется,
	// поэтому текст пишем целиком: обрезанное письмо прочитать нельзя. Адрес
	// клиента нужен, чтобы отличить поток однотипных сообщений от разных людей.
	zap.L().Info(
		"contact message received",
		zap.String("email", email),
		zap.String("subject", subject),
		zap.String("message", message),
		zap.String("client_ip", ctx.RealIP()),
	)

	return ctx.JSON(http.StatusAccepted, spec.ContactMessageResponse{
		Message: "Message accepted. We will reply to the address you provided.",
	})
}

// validateContactMessage приводит поля к виду, пригодному для отправки письма, и
// возвращает причину отказа текстом: посетителю в форме показывается именно она,
// поэтому формулировка должна объяснять, что поправить.
func validateContactMessage(req spec.ContactMessageRequest) (email, subject, message string, err error) {
	email = strings.TrimSpace(string(req.Email))
	subject = strings.TrimSpace(req.Subject)
	message = strings.TrimSpace(req.Message)

	// Разбор по RFC 5322 принимает и запись с именем («Ann <ann@example.com>»),
	// но в поле «от кого» ждём один адрес: сравнение с разобранным отсекает
	// остальное, не заводя своей регулярки.
	addr, parseErr := mail.ParseAddress(email)
	if parseErr != nil || addr.Address != email || len(email) > contactEmailMaxLen {
		return "", "", "", errContactEmailInvalid
	}
	if n := len([]rune(subject)); n < contactSubjectMinLen || n > contactSubjectMaxLen {
		return "", "", "", errContactSubjectLength
	}
	if n := len([]rune(message)); n < contactMessageMinLen || n > contactMessageMaxLen {
		return "", "", "", errContactMessageLength
	}
	return email, subject, message, nil
}
