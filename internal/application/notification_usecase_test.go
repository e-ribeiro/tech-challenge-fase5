package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"video-processor/internal/adapter/memory"
	"video-processor/internal/application"
	"video-processor/internal/port"
)

func TestNotificationUseCase_HandleNotification_Success(t *testing.T) {
	notifier := memory.NewMemoryNotifier()
	uc := application.NewNotificationUseCase(notifier)

	ctx := context.Background()
	msg := &port.NotificationMessage{
		JobID:     "job-999",
		UserID:    "user-1",
		UserEmail: "cliente@fiap.com",
		Type:      "FAILED",
		Title:     "Erro no processamento do seu vídeo",
		Message:   "O arquivo enviado não pôde ser convertido.",
		CreatedAt: time.Now(),
	}

	err := uc.HandleNotification(ctx, msg)
	require.NoError(t, err)

	sent := notifier.GetSentNotifications()
	require.Len(t, sent, 1)
	assert.Equal(t, "cliente@fiap.com", sent[0].ToEmail)
	assert.Contains(t, sent[0].Subject, "[FIAP X]")
	assert.Contains(t, sent[0].Body, "job-999")
	assert.Contains(t, sent[0].Body, "O arquivo enviado não pôde ser convertido.")
}
