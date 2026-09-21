package mailer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"video-processor/internal/adapter/mailer"
	"video-processor/internal/port"
)

func TestSMTPMailer_SendNotification_ConnectionFail(t *testing.T) {
	// Testa o comportamento quando a porta SMTP não está acessível
	m := mailer.NewSMTPMailer("127.0.0.1", 59999, "", "", "no-reply@fiapx.com")

	err := m.SendNotification(context.Background(), &port.NotificationRequest{
		ToEmail: "teste@fiap.com",
		Subject: "Teste",
		Body:    "Corpo de teste",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "falha ao enviar e-mail via SMTP")
}
