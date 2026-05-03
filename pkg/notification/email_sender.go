package notification

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

type EmailSender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewEmailSender(
	host string,
	port int,
	username string,
	password string,
	from string,
) *EmailSender {
	return &EmailSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (e *EmailSender) Channel() Channel {
	return ChannelEmail
}

func (e *EmailSender) Send(ctx context.Context, n Notification) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	em := &email.Email{
		From:    e.from,
		To:      []string{n.To},
		Subject: n.Subject,
	}

	if n.HTMLBody != "" {
		em.HTML = []byte(n.HTMLBody)
	}

	if n.TextBody != "" {
		em.Text = []byte(n.TextBody)
	}

	addr := fmt.Sprintf("%s:%d", e.host, e.port)
	auth := smtp.PlainAuth("", e.username, e.password, e.host)

	return em.Send(addr, auth)
}
