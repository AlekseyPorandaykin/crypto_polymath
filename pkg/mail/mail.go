package mail

type Mailer interface {
	Send(subject, body string) error
}
