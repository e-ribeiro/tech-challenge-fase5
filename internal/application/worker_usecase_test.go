package application_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"video-processor/internal/adapter/memory"
	"video-processor/internal/adapter/processor"
	"video-processor/internal/application"
	"video-processor/internal/domain"
	"video-processor/internal/port"
)

func TestWorkerUseCase_ProcessJob_Success(t *testing.T) {
	repo := memory.NewMemoryVideoJobRepository()
	storage := memory.NewMemoryStorage()
	queue := memory.NewMemoryQueue(10)
	mockProcessor := processor.NewMockVideoProcessor(25, false)

	uc := application.NewWorkerUseCase(repo, storage, mockProcessor, queue)

	ctx := context.Background()

	// Prepara arquivo de vídeo no storage
	videoKey := "videos/test_video.mp4"
	_ = storage.Upload(ctx, videoKey, strings.NewReader("video-bytes"), "video/mp4")

	job, err := domain.NewVideoJob("job-abc", "user-1", "test_video.mp4", videoKey)
	require.NoError(t, err)
	_ = repo.Create(ctx, job)

	msg := &port.VideoProcessMessage{
		JobID:        job.ID,
		UserID:       job.UserID,
		UserEmail:    "investidor@fiap.com",
		OriginalName: job.OriginalName,
		VideoKey:     job.VideoKey,
		CreatedAt:    time.Now(),
	}

	err = uc.ProcessJob(ctx, msg)
	require.NoError(t, err)

	// Verifica se o job foi atualizado para COMPLETED
	updatedJob, err := repo.GetByID(ctx, "job-abc")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusCompleted, updatedJob.Status)
	assert.Equal(t, 25, updatedJob.FrameCount)
	assert.NotEmpty(t, updatedJob.ZipKey)

	// Verifica se o ZIP está no storage
	zipReader, err := storage.Download(ctx, updatedJob.ZipKey)
	require.NoError(t, err)
	assert.NotNil(t, zipReader)

	// Verifica se evento de notificação de conclusão foi gerado
	notifications := queue.GetPublishedNotificationMessages()
	require.Len(t, notifications, 1)
	assert.Equal(t, "COMPLETED", notifications[0].Type)
	assert.Equal(t, "investidor@fiap.com", notifications[0].UserEmail)
}

func TestWorkerUseCase_ProcessJob_FailureAndNotification(t *testing.T) {
	repo := memory.NewMemoryVideoJobRepository()
	storage := memory.NewMemoryStorage()
	queue := memory.NewMemoryQueue(10)
	mockProcessor := processor.NewMockVideoProcessor(0, true) // Força erro no processamento

	uc := application.NewWorkerUseCase(repo, storage, mockProcessor, queue)

	ctx := context.Background()
	videoKey := "videos/corrupted.mp4"
	_ = storage.Upload(ctx, videoKey, strings.NewReader("corrupted-bytes"), "video/mp4")

	job, err := domain.NewVideoJob("job-fail", "user-1", "corrupted.mp4", videoKey)
	require.NoError(t, err)
	_ = repo.Create(ctx, job)

	msg := &port.VideoProcessMessage{
		JobID:        job.ID,
		UserID:       job.UserID,
		UserEmail:    "investidor@fiap.com",
		OriginalName: job.OriginalName,
		VideoKey:     job.VideoKey,
		CreatedAt:    time.Now(),
	}

	err = uc.ProcessJob(ctx, msg)
	assert.Error(t, err)

	// Verifica se o job foi marcado como FAILED
	updatedJob, err := repo.GetByID(ctx, "job-fail")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusFailed, updatedJob.Status)
	assert.Contains(t, updatedJob.ErrorMessage, "falha simulada")

	// Verifica se evento de notificação de erro foi gerado para o usuário
	notifications := queue.GetPublishedNotificationMessages()
	require.Len(t, notifications, 1)
	assert.Equal(t, "FAILED", notifications[0].Type)
	assert.Equal(t, "investidor@fiap.com", notifications[0].UserEmail)
	assert.Contains(t, notifications[0].Message, "falha ao processar")
}
