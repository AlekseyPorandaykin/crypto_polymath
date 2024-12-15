package mail

import (
	"context"
	"errors"
	"fmt"
	"github.com/labstack/gommon/log"
	"net/smtp"
	"sync"
	"time"
)

var ErrSMTPClosed = errors.New("smtp is closed")

var timeoutClose = 10 * time.Second

type message struct {
	from    []string
	to      []string
	subject string
	body    string
}

type Smtp struct {
	username string
	password string
	host     string
	port     uint
	from     string
	to       string

	messageCh chan message
	start     sync.Once
	closed    bool
}

func NewSmtp() *Smtp {
	s := &Smtp{}
	s.start.Do(func() {
		go s.worker()
	})
	return s
}

func (s *Smtp) Send(from, to []string, subject, body string) error {
	if s.closed {
		return ErrSMTPClosed
	}
	s.messageCh <- message{
		from:    from,
		to:      to,
		subject: subject,
		body:    body,
	}
	return nil
}

func (s *Smtp) worker() {
	for msg := range s.messageCh {
		if err := s.sendMail(msg); err != nil {
			log.Error("error send mail: ", err)
		}
	}
}

func (s *Smtp) sendMail(msg message) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	addr := s.host
	if s.port > 0 {
		addr = fmt.Sprintf("%s:%d", s.host, s.port)
	}
	body := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", msg.to, msg.subject, msg.body))
	return smtp.SendMail(addr, auth, s.from, msg.to, body)
}

func (s *Smtp) Close() {
	s.closed = true
	ctx, cancel := context.WithTimeout(context.Background(), timeoutClose)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if len(s.messageCh) == 0 {
				close(s.messageCh)
				cancel()
			}
		}
	}
}
