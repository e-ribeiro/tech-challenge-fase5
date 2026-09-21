package memory

import (
	"context"
	"sync"

	"video-processor/internal/port"
)

type MemoryNotifier struct {
	mu            sync.RWMutex
	notifications []*port.NotificationRequest
}

func NewMemoryNotifier() *MemoryNotifier {
	return &MemoryNotifier{
		notifications: make([]*port.NotificationRequest, 0),
	}
}

func (n *MemoryNotifier) SendNotification(ctx context.Context, req *port.NotificationRequest) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	copyReq := *req
	n.notifications = append(n.notifications, &copyReq)
	return nil
}

func (n *MemoryNotifier) GetSentNotifications() []*port.NotificationRequest {
	n.mu.RLock()
	defer n.mu.RUnlock()

	copied := make([]*port.NotificationRequest, len(n.notifications))
	copy(copied, n.notifications)
	return copied
}

var _ port.Notifier = (*MemoryNotifier)(nil)
