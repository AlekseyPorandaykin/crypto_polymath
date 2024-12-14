package logger

import (
	"fmt"
	"go.uber.org/zap/zapcore"
	"log"
	"net/smtp"
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
		body += fmt.Sprintf("%s: %s\n", field.Key, field.String)
	}
	body += fmt.Sprintf("\n\nStacktrace:\n%s", entry.Stack)
	subject := fmt.Sprintf("%s [%s]: %s", entry.LoggerName, entry.Level.String(), entry.Time.String())
	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	to := []string{m.to}

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", m.to, subject, body))
	addr := m.host
	if m.port > 0 {
		addr = fmt.Sprintf("%s:%d", m.host, m.port)
	}
	err := smtp.SendMail(addr, auth, m.from, to, msg)
	if err != nil {
		log.Printf("Couldn't send email: %v", err)
	}
	return nil
}

func (m *MailZapCore) Sync() error {
	return nil
}
