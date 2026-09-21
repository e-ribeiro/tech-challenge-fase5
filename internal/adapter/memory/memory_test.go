package memory_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"video-processor/internal/adapter/memory"
	"video-processor/internal/domain"
	"video-processor/internal/port"
)

func TestMemoryUserRepository(t *testing.T) {
	repo := memory.NewMemoryUserRepository()
	ctx := context.Background()

	user, err := domain.NewUser("u1", "Nome", "teste@fiap.com", "senha123")
	require.NoError(t, err)

	// Create
	err = repo.Create(ctx, user)
	require.NoError(t, err)

	// Create Duplicate
	err = repo.Create(ctx, user)
	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)

	// GetByID
	found, err := repo.GetByID(ctx, "u1")
	require.NoError(t, err)
	assert.Equal(t, "Nome", found.Name)

	_, err = repo.GetByID(ctx, "invalido")
	assert.ErrorIs(t, err, domain.ErrUserNotFound)

	// GetByEmail
	foundEmail, err := repo.GetByEmail(ctx, "teste@fiap.com")
	require.NoError(t, err)
	assert.Equal(t, "u1", foundEmail.ID)

	_, err = repo.GetByEmail(ctx, "outro@fiap.com")
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestMemoryVideoJobRepository(t *testing.T) {
	repo := memory.NewMemoryVideoJobRepository()
	ctx := context.Background()

	job, err := domain.NewVideoJob("j1", "u1", "vid.mp4", "videos/vid.mp4")
	require.NoError(t, err)

	// Create
	err = repo.Create(ctx, job)
	require.NoError(t, err)

	// GetByID
	found, err := repo.GetByID(ctx, "j1")
	require.NoError(t, err)
	assert.Equal(t, "vid.mp4", found.OriginalName)

	_, err = repo.GetByID(ctx, "nao-existe")
	assert.ErrorIs(t, err, domain.ErrVideoNotFound)

	// GetByUserID
	list, err := repo.GetByUserID(ctx, "u1")
	require.NoError(t, err)
	assert.Len(t, list, 1)

	listEmpty, err := repo.GetByUserID(ctx, "outro-user")
	require.NoError(t, err)
	assert.Empty(t, listEmpty)

	// Update
	job.MarkCompleted("outputs/frames.zip", 10)
	err = repo.Update(ctx, job)
	require.NoError(t, err)

	jobInexistente, _ := domain.NewVideoJob("inexistente", "u1", "vid.mp4", "v.mp4")
	err = repo.Update(ctx, jobInexistente)
	assert.ErrorIs(t, err, domain.ErrVideoNotFound)
}

func TestMemoryStorage(t *testing.T) {
	s := memory.NewMemoryStorage()
	ctx := context.Background()

	err := s.Upload(ctx, "k1", strings.NewReader("dados"), "text/plain")
	require.NoError(t, err)

	r, err := s.Download(ctx, "k1")
	require.NoError(t, err)
	_ = r.Close()

	_, err = s.Download(ctx, "inexistente")
	assert.Error(t, err)

	url, err := s.GetFileURL(ctx, "k1")
	require.NoError(t, err)
	assert.Contains(t, url, "k1")

	_, err = s.GetFileURL(ctx, "inexistente")
	assert.Error(t, err)

	err = s.Delete(ctx, "k1")
	require.NoError(t, err)
}

func TestMemoryQueueAndNotifier(t *testing.T) {
	q := memory.NewMemoryQueue(10)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	vMsg := &port.VideoProcessMessage{JobID: "j1", OriginalName: "vid.mp4"}
	err := q.PublishVideoProcess(ctx, vMsg)
	require.NoError(t, err)

	nMsg := &port.NotificationMessage{JobID: "j1", Type: "COMPLETED"}
	err = q.PublishNotification(ctx, nMsg)
	require.NoError(t, err)

	assert.Len(t, q.GetPublishedVideoMessages(), 1)
	assert.Len(t, q.GetPublishedNotificationMessages(), 1)

	// Consumidores com contexto cancelado
	go func() {
		_ = q.ConsumeVideoProcess(ctx, func(c context.Context, m *port.VideoProcessMessage) error {
			return nil
		})
	}()

	go func() {
		_ = q.ConsumeNotification(ctx, func(c context.Context, m *port.NotificationMessage) error {
			return nil
		})
	}()

	time.Sleep(100 * time.Millisecond)
	_ = q.Close()

	// Notifier
	notifier := memory.NewMemoryNotifier()
	err = notifier.SendNotification(ctx, &port.NotificationRequest{
		ToEmail: "user@fiap.com",
		Subject: "Ola",
		Body:    "Corpo",
	})
	require.NoError(t, err)
	assert.Len(t, notifier.GetSentNotifications(), 1)
}
