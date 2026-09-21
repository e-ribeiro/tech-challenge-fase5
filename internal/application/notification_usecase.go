package application

import (
	"context"
	"fmt"

	"video-processor/internal/port"
)

type NotificationUseCase struct {
	notifier port.Notifier
}

func NewNotificationUseCase(notifier port.Notifier) *NotificationUseCase {
	return &NotificationUseCase{
		notifier: notifier,
	}
}

func (uc *NotificationUseCase) HandleNotification(ctx context.Context, msg *port.NotificationMessage) error {
	if msg.UserEmail == "" {
		return nil
	}

	body := fmt.Sprintf(
		"Prezado usuário,\n\n%s\n\nIdentificador do job: %s\nData: %s\n\nAtenciosamente,\nEquipe FIAP X",
		msg.Message,
		msg.JobID,
		msg.CreatedAt.Format("02/01/2006 15:04:05"),
	)

	req := &port.NotificationRequest{
		ToEmail: msg.UserEmail,
		Subject: fmt.Sprintf("[FIAP X] %s", msg.Title),
		Body:    body,
	}

	return uc.notifier.SendNotification(ctx, req)
}
