package memory

import (
	"context"
	"sync"

	"video-processor/internal/port"
)

type MemoryQueue struct {
	mu         sync.Mutex
	videoChan  chan *port.VideoProcessMessage
	notifyChan chan *port.NotificationMessage
	videoMsgs  []*port.VideoProcessMessage
	notifyMsgs []*port.NotificationMessage
	closed     bool
}

func NewMemoryQueue(bufferSize int) *MemoryQueue {
	if bufferSize <= 0 {
		bufferSize = 100
	}
	return &MemoryQueue{
		videoChan:  make(chan *port.VideoProcessMessage, bufferSize),
		notifyChan: make(chan *port.NotificationMessage, bufferSize),
		videoMsgs:  make([]*port.VideoProcessMessage, 0),
		notifyMsgs: make([]*port.NotificationMessage, 0),
	}
}

func (q *MemoryQueue) PublishVideoProcess(ctx context.Context, msg *port.VideoProcessMessage) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.videoMsgs = append(q.videoMsgs, msg)
	select {
	case q.videoChan <- msg:
	default:
	}
	return nil
}

func (q *MemoryQueue) PublishNotification(ctx context.Context, msg *port.NotificationMessage) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.notifyMsgs = append(q.notifyMsgs, msg)
	select {
	case q.notifyChan <- msg:
	default:
	}
	return nil
}

func (q *MemoryQueue) ConsumeVideoProcess(ctx context.Context, handler func(ctx context.Context, msg *port.VideoProcessMessage) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-q.videoChan:
			if !ok {
				return nil
			}
			_ = handler(ctx, msg)
		}
	}
}

func (q *MemoryQueue) ConsumeNotification(ctx context.Context, handler func(ctx context.Context, msg *port.NotificationMessage) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-q.notifyChan:
			if !ok {
				return nil
			}
			_ = handler(ctx, msg)
		}
	}
}

func (q *MemoryQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.closed {
		close(q.videoChan)
		close(q.notifyChan)
		q.closed = true
	}
	return nil
}

func (q *MemoryQueue) GetPublishedVideoMessages() []*port.VideoProcessMessage {
	q.mu.Lock()
	defer q.mu.Unlock()
	copied := make([]*port.VideoProcessMessage, len(q.videoMsgs))
	copy(copied, q.videoMsgs)
	return copied
}

func (q *MemoryQueue) GetPublishedNotificationMessages() []*port.NotificationMessage {
	q.mu.Lock()
	defer q.mu.Unlock()
	copied := make([]*port.NotificationMessage, len(q.notifyMsgs))
	copy(copied, q.notifyMsgs)
	return copied
}

var _ port.QueueProducer = (*MemoryQueue)(nil)
var _ port.QueueConsumer = (*MemoryQueue)(nil)
