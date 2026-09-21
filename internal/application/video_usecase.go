package application

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"video-processor/internal/domain"
	"video-processor/internal/port"
)

type VideoUseCase struct {
	videoRepo port.VideoJobRepository
	storage   port.Storage
	producer  port.QueueProducer
}

func NewVideoUseCase(videoRepo port.VideoJobRepository, storage port.Storage, producer port.QueueProducer) *VideoUseCase {
	return &VideoUseCase{
		videoRepo: videoRepo,
		storage:   storage,
		producer:  producer,
	}
}

type UploadVideoInput struct {
	UserID      string
	UserEmail   string
	Filename    string
	ContentType string
	Content     io.Reader
}

func (uc *VideoUseCase) UploadVideo(ctx context.Context, input UploadVideoInput) (*domain.VideoJob, error) {
	if !domain.IsValidVideoExtension(input.Filename) {
		return nil, domain.ErrUnsupportedVideoType
	}

	jobID := uuid.New().String()
	ext := filepath.Ext(input.Filename)
	timestamp := time.Now().Format("20060102_150405")
	videoKey := fmt.Sprintf("videos/%s_%s%s", timestamp, jobID, ext)

	// Salva no storage (S3/MinIO ou local)
	if err := uc.storage.Upload(ctx, videoKey, input.Content, input.ContentType); err != nil {
		return nil, fmt.Errorf("falha ao armazenar vídeo: %w", err)
	}

	job, err := domain.NewVideoJob(jobID, input.UserID, input.Filename, videoKey)
	if err != nil {
		return nil, err
	}

	if err := uc.videoRepo.Create(ctx, job); err != nil {
		// Se falhar no banco, tenta remover do storage
		_ = uc.storage.Delete(ctx, videoKey)
		return nil, fmt.Errorf("falha ao persistir job de vídeo: %w", err)
	}

	// Enfileira mensagem para processamento assíncrono pelos workers
	msg := &port.VideoProcessMessage{
		JobID:        job.ID,
		UserID:       job.UserID,
		UserEmail:    input.UserEmail,
		OriginalName: job.OriginalName,
		VideoKey:     job.VideoKey,
		CreatedAt:    job.CreatedAt,
	}

	if err := uc.producer.PublishVideoProcess(ctx, msg); err != nil {
		job.MarkFailed("falha ao enfileirar job para processamento")
		_ = uc.videoRepo.Update(ctx, job)
		return nil, fmt.Errorf("falha ao enviar para a fila: %w", err)
	}

	return job, nil
}

func (uc *VideoUseCase) ListUserVideos(ctx context.Context, userID string) ([]*domain.VideoJob, error) {
	return uc.videoRepo.GetByUserID(ctx, userID)
}

func (uc *VideoUseCase) GetVideoByID(ctx context.Context, userID, videoID string) (*domain.VideoJob, error) {
	job, err := uc.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, domain.ErrVideoNotFound
	}

	if job.UserID != userID {
		return nil, domain.ErrVideoForbidden
	}

	return job, nil
}

func (uc *VideoUseCase) GetVideoDownload(ctx context.Context, userID, videoID string) (io.ReadCloser, string, error) {
	job, err := uc.GetVideoByID(ctx, userID, videoID)
	if err != nil {
		return nil, "", err
	}

	if job.Status != domain.StatusCompleted {
		return nil, "", domain.ErrVideoNotReady
	}

	if job.ZipKey == "" {
		return nil, "", domain.ErrVideoNotFound
	}

	reader, err := uc.storage.Download(ctx, job.ZipKey)
	if err != nil {
		return nil, "", fmt.Errorf("erro ao recuperar arquivo compactado: %w", err)
	}

	filename := filepath.Base(job.ZipKey)
	return reader, filename, nil
}
