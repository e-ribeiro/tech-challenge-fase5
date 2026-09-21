package application_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"video-processor/internal/adapter/memory"
	"video-processor/internal/application"
	"video-processor/internal/domain"
)

func TestVideoUseCase_UploadVideo_Success(t *testing.T) {
	repo := memory.NewMemoryVideoJobRepository()
	storage := memory.NewMemoryStorage()
	queue := memory.NewMemoryQueue(10)

	uc := application.NewVideoUseCase(repo, storage, queue)

	ctx := context.Background()
	content := strings.NewReader("fake video binary content")

	job, err := uc.UploadVideo(ctx, application.UploadVideoInput{
		UserID:      "user-123",
		UserEmail:   "user@fiap.com",
		Filename:    "apresentacao.mp4",
		ContentType: "video/mp4",
		Content:     content,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, job.ID)
	assert.Equal(t, "user-123", job.UserID)
	assert.Equal(t, "apresentacao.mp4", job.OriginalName)
	assert.Equal(t, domain.StatusPending, job.Status)

	// Valida persistência no repositório
	saved, err := repo.GetByID(ctx, job.ID)
	require.NoError(t, err)
	assert.Equal(t, job.ID, saved.ID)

	// Valida publicação na fila de mensageria
	messages := queue.GetPublishedVideoMessages()
	require.Len(t, messages, 1)
	assert.Equal(t, job.ID, messages[0].JobID)
	assert.Equal(t, "user@fiap.com", messages[0].UserEmail)
}

func TestVideoUseCase_UploadVideo_InvalidExtension(t *testing.T) {
	repo := memory.NewMemoryVideoJobRepository()
	storage := memory.NewMemoryStorage()
	queue := memory.NewMemoryQueue(10)

	uc := application.NewVideoUseCase(repo, storage, queue)

	ctx := context.Background()
	content := strings.NewReader("pdf content")

	_, err := uc.UploadVideo(ctx, application.UploadVideoInput{
		UserID:      "user-123",
		UserEmail:   "user@fiap.com",
		Filename:    "relatorio.pdf",
		ContentType: "application/pdf",
		Content:     content,
	})

	assert.ErrorIs(t, err, domain.ErrUnsupportedVideoType)
}

func TestVideoUseCase_ListUserVideos(t *testing.T) {
	repo := memory.NewMemoryVideoJobRepository()
	storage := memory.NewMemoryStorage()
	queue := memory.NewMemoryQueue(10)

	uc := application.NewVideoUseCase(repo, storage, queue)
	ctx := context.Background()

	// Cria jobs para user-1 e user-2
	_, _ = uc.UploadVideo(ctx, application.UploadVideoInput{
		UserID: "user-1", UserEmail: "u1@fiap.com", Filename: "vid1.mp4", Content: strings.NewReader("1"),
	})
	_, _ = uc.UploadVideo(ctx, application.UploadVideoInput{
		UserID: "user-1", UserEmail: "u1@fiap.com", Filename: "vid2.mov", Content: strings.NewReader("2"),
	})
	_, _ = uc.UploadVideo(ctx, application.UploadVideoInput{
		UserID: "user-2", UserEmail: "u2@fiap.com", Filename: "vid3.mp4", Content: strings.NewReader("3"),
	})

	list1, err := uc.ListUserVideos(ctx, "user-1")
	require.NoError(t, err)
	assert.Len(t, list1, 2)

	list2, err := uc.ListUserVideos(ctx, "user-2")
	require.NoError(t, err)
	assert.Len(t, list2, 1)
}

func TestVideoUseCase_GetVideoByID_And_Download(t *testing.T) {
	repo := memory.NewMemoryVideoJobRepository()
	storage := memory.NewMemoryStorage()
	queue := memory.NewMemoryQueue(10)

	uc := application.NewVideoUseCase(repo, storage, queue)
	ctx := context.Background()

	job, err := uc.UploadVideo(ctx, application.UploadVideoInput{
		UserID: "user-1", UserEmail: "u1@fiap.com", Filename: "vid1.mp4", Content: strings.NewReader("data"),
	})
	require.NoError(t, err)

	// Consulta de outro usuário deve dar erro de permissão
	_, err = uc.GetVideoByID(ctx, "user-hacker", job.ID)
	assert.ErrorIs(t, err, domain.ErrVideoForbidden)

	// Download antes de concluir deve falhar
	_, _, err = uc.GetVideoDownload(ctx, "user-1", job.ID)
	assert.ErrorIs(t, err, domain.ErrVideoNotReady)

	// Simula conclusão
	zipKey := "outputs/frames_completed.zip"
	_ = storage.Upload(ctx, zipKey, bytes.NewReader([]byte("fake zip content")), "application/zip")
	job.MarkCompleted(zipKey, 15)
	_ = repo.Update(ctx, job)

	// Download agora deve ser bem sucedido
	reader, filename, err := uc.GetVideoDownload(ctx, "user-1", job.ID)
	require.NoError(t, err)
	assert.Equal(t, "frames_completed.zip", filename)

	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "fake zip content", string(content))
}
