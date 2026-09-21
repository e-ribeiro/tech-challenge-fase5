package mailer

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"video-processor/internal/port"
)

type SMTPMailer struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewSMTPMailer(host string, port int, username, password, from string) *SMTPMailer {
	if from == "" {
		from = "no-reply@fiapx.com"
	}
	return &SMTPMailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (m *SMTPMailer) SendNotification(ctx context.Context, req *port.NotificationRequest) error {
	addr := fmt.Sprintf("%s:%d", m.host, m.port)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.from,
		req.ToEmail,
		req.Subject,
		req.Body,
	)

	var auth smtp.Auth
	if m.username != "" && m.password != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}

	to := []string{strings.TrimSpace(req.ToEmail)}
	err := smtp.SendMail(addr, auth, m.from, to, []byte(msg))
	if err != nil {
		return fmt.Errorf("falha ao enviar e-mail via SMTP (%s): %w", addr, err)
	}

	return nil
}

var _ port.Notifier = (*SMTPMailer)(nil)
