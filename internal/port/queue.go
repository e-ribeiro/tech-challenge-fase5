package port

import (
	"context"
	"time"
)

type VideoProcessMessage struct {
	JobID        string    `json:"job_id"`
	UserID       string    `json:"user_id"`
	UserEmail    string    `json:"user_email"`
	OriginalName string    `json:"original_name"`
	VideoKey     string    `json:"video_key"`
	CreatedAt    time.Time `json:"created_at"`
}

type NotificationMessage struct {
	JobID     string    `json:"job_id"`
	UserID    string    `json:"user_id"`
	UserEmail string    `json:"user_email"`
	Type      string    `json:"type"` // "FAILED" ou "COMPLETED"
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type QueueProducer interface {
	PublishVideoProcess(ctx context.Context, msg *VideoProcessMessage) error
	PublishNotification(ctx context.Context, msg *NotificationMessage) error
	Close() error
}

type QueueConsumer interface {
	ConsumeVideoProcess(ctx context.Context, handler func(ctx context.Context, msg *VideoProcessMessage) error) error
	ConsumeNotification(ctx context.Context, handler func(ctx context.Context, msg *NotificationMessage) error) error
	Close() error
}
