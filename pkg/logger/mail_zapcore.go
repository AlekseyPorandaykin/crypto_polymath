package logger

import (
	"context"
	"fmt"
	mailslurp "github.com/mailslurp/mailslurp-client-go"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

var _ zapcore.Core = (*MailZapCore)(nil)

type MailZapCore struct {
	fields []zapcore.Field
	level  zapcore.Level

	// SMTP settings
	username string
	password string
	host     string
	port     uint
	from     string
	to       string

	client *mailslurp.APIClient
}

func NewMailZapCore(level zapcore.Level, username, password, host string, port uint, from, to string) *MailZapCore {
	return &MailZapCore{
		fields:   make([]zapcore.Field, 0),
		level:    level,
		username: username,
		password: password,
		host:     host,
		port:     port,
		from:     from,
		to:       to,
		client:   mailslurp.NewAPIClient(mailslurp.NewConfiguration()),
	}
}

func (m *MailZapCore) Enabled(level zapcore.Level) bool {
	return m.level.Enabled(level)
}

func (m *MailZapCore) With(fields []zapcore.Field) zapcore.Core {
	return &MailZapCore{
		fields:   fields,
		level:    m.level,
		username: m.username,
		password: m.password,
		host:     m.host,
		port:     m.port,
		from:     m.from,
		to:       m.to,
	}
}

func (m *MailZapCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if m.Enabled(ent.Level) {
		return ce.AddCore(ent, m)
	}
	return ce
}

func (m *MailZapCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Create a new email message
	message := `
Logger Name: %s
Level: %s
Time: %s
Message: %s

`
	body := fmt.Sprintf(message, entry.LoggerName, entry.Level.String(), entry.Time.String(), entry.Message)
	fields = append(fields, m.fields...)
	if len(fields) > 0 {
		body += "\nFields:\n"
	}
	for _, field := range fields {
		if field.String != "" {
			body += fmt.Sprintf("%s: %s\n", field.Key, field.String)
		}
		if field.Integer > 0 {
			body += fmt.Sprintf("%s: %d\n", field.Key, field.Integer)
		}
		if field.Interface != nil {
			body += fmt.Sprintf("%s: %v\n", field.Key, field.Interface)
		}
	}
	body += fmt.Sprintf("\n\nStacktrace:\n%s", entry.Stack)
	subject := fmt.Sprintf("%s [%s]: %s", entry.LoggerName, entry.Level.String(), entry.Time.String())
	if err := m.sent([]string{m.to}, subject, body); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Couldn't send email (%s:%s)", subject, body))
	}
	return nil
}

func (m *MailZapCore) sent(to []string, subject, body string) error {
	options := mailslurp.SendEmailOptions{
		To:      &to,
		Subject: &subject,
		Body:    &body,
	}
	ctx := context.WithValue(
		context.Background(),
		mailslurp.ContextAPIKey,
		mailslurp.APIKey{Key: viper.GetString("mailslurp.api_key")})
	_, err := m.client.InboxControllerApi.SendEmail(ctx, viper.GetString("mailslurp.inbox_id"), options)
	return err
}

func (m *MailZapCore) Sync() error {
	return nil
}
