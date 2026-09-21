package application

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"video-processor/internal/domain"
	"video-processor/internal/port"
)

type WorkerUseCase struct {
	videoRepo port.VideoJobRepository
	storage   port.Storage
	processor port.VideoProcessor
	producer  port.QueueProducer
}

func NewWorkerUseCase(
	videoRepo port.VideoJobRepository,
	storage port.Storage,
	processor port.VideoProcessor,
	producer port.QueueProducer,
) *WorkerUseCase {
	return &WorkerUseCase{
		videoRepo: videoRepo,
		storage:   storage,
		processor: processor,
		producer:  producer,
	}
}

func (uc *WorkerUseCase) ProcessJob(ctx context.Context, msg *port.VideoProcessMessage) error {
	job, err := uc.videoRepo.GetByID(ctx, msg.JobID)
	if err != nil {
		return fmt.Errorf("job %s não encontrado no banco: %w", msg.JobID, err)
	}

	job.MarkProcessing()
	_ = uc.videoRepo.Update(ctx, job)

	// Cria diretório temporário de trabalho para esta tarefa
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("fiapx_%s_*", job.ID))
	if err != nil {
		uc.failJob(ctx, job, msg, fmt.Sprintf("falha ao criar pasta temporária: %v", err))
		return err
	}
	defer os.RemoveAll(tempDir)

	// 1. Baixar o vídeo do Object Storage para o disco local do worker
	localVideoPath := filepath.Join(tempDir, filepath.Base(job.VideoKey))
	if err := uc.downloadToLocal(ctx, job.VideoKey, localVideoPath); err != nil {
		uc.failJob(ctx, job, msg, fmt.Sprintf("falha ao baixar vídeo do storage: %v", err))
		return err
	}

	// 2. Processar vídeo com ffmpeg e compactar os frames em .zip
	result, err := uc.processor.ProcessVideoToZip(ctx, localVideoPath, 1)
	if err != nil {
		uc.failJob(ctx, job, msg, fmt.Sprintf("falha na extração de frames: %v", err))
		return err
	}

	// 3. Fazer upload do arquivo .zip gerado para o Object Storage
	zipFile, err := os.Open(result.ZipLocalPath)
	if err != nil {
		uc.failJob(ctx, job, msg, fmt.Sprintf("falha ao abrir zip gerado: %v", err))
		return err
	}
	defer zipFile.Close()

	timestamp := time.Now().Format("20060102_150405")
	zipKey := fmt.Sprintf("outputs/frames_%s_%s.zip", timestamp, job.ID)

	if err := uc.storage.Upload(ctx, zipKey, zipFile, "application/zip"); err != nil {
		uc.failJob(ctx, job, msg, fmt.Sprintf("falha ao salvar zip no storage: %v", err))
		return err
	}

	// 4. Marcar job como COMPLETED no banco de dados
	job.MarkCompleted(zipKey, result.FrameCount)
	if err := uc.videoRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("falha ao atualizar status de conclusão do job: %w", err)
	}

	// 5. Publicar evento de conclusão com sucesso
	if uc.producer != nil && msg.UserEmail != "" {
		_ = uc.producer.PublishNotification(ctx, &port.NotificationMessage{
			JobID:     job.ID,
			UserID:    job.UserID,
			UserEmail: msg.UserEmail,
			Type:      "COMPLETED",
			Title:     "Processamento de vídeo concluído com sucesso!",
			Message:   fmt.Sprintf("Seu vídeo '%s' foi processado com sucesso. Foram extraídos %d frames.", job.OriginalName, result.FrameCount),
			CreatedAt: time.Now().UTC(),
		})
	}

	return nil
}

func (uc *WorkerUseCase) downloadToLocal(ctx context.Context, storageKey, localPath string) error {
	reader, err := uc.storage.Download(ctx, storageKey)
	if err != nil {
		return err
	}
	defer reader.Close()

	outFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, reader)
	return err
}

func (uc *WorkerUseCase) failJob(ctx context.Context, job *domain.VideoJob, msg *port.VideoProcessMessage, reason string) {
	job.MarkFailed(reason)
	_ = uc.videoRepo.Update(ctx, job)

	if uc.producer != nil && msg.UserEmail != "" {
		_ = uc.producer.PublishNotification(ctx, &port.NotificationMessage{
			JobID:     job.ID,
			UserID:    job.UserID,
			UserEmail: msg.UserEmail,
			Type:      "FAILED",
			Title:     "Erro no processamento do seu vídeo",
			Message:   fmt.Sprintf("Houve uma falha ao processar o vídeo '%s'. Motivo: %s", job.OriginalName, reason),
			CreatedAt: time.Now().UTC(),
		})
	}
}
