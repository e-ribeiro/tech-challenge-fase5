package port

import (
	"context"
)

type NotificationRequest struct {
	ToEmail string
	Subject string
	Body    string
}

type Notifier interface {
	SendNotification(ctx context.Context, req *NotificationRequest) error
}
