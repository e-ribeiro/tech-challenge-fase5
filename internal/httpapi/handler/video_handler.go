package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"video-processor/internal/application"
	"video-processor/internal/domain"
	"video-processor/internal/httpapi/dto"
	"video-processor/internal/httpapi/middleware"
)

type VideoHandler struct {
	videoUC *application.VideoUseCase
}

func NewVideoHandler(videoUC *application.VideoUseCase) *VideoHandler {
	return &VideoHandler{videoUC: videoUC}
}

func (h *VideoHandler) Upload(c *gin.Context) {
	userID, _ := c.Get(middleware.CtxUserIDKey)
	userEmail, _ := c.Get(middleware.CtxUserEmailKey)

	fileHeader, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "arquivo de vídeo não fornecido: " + err.Error()})
		return
	}

	if !domain.IsValidVideoExtension(fileHeader.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "formato de arquivo não suportado. Formatos aceitos: mp4, avi, mov, mkv, wmv, flv, webm",
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao abrir arquivo: " + err.Error()})
		return
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	job, err := h.videoUC.UploadVideo(c.Request.Context(), application.UploadVideoInput{
		UserID:      userID.(string),
		UserEmail:   userEmail.(string),
		Filename:    fileHeader.Filename,
		ContentType: contentType,
		Content:     file,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao processar requisição de upload: " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, dto.UploadVideoResponse{
		JobID:   job.ID,
		Status:  string(job.Status),
		Message: "Vídeo recebido e enfileirado com sucesso para processamento.",
	})
}

func (h *VideoHandler) List(c *gin.Context) {
	userID, _ := c.Get(middleware.CtxUserIDKey)

	jobs, err := h.videoUC.ListUserVideos(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao listar vídeos: " + err.Error()})
		return
	}

	resp := make([]*dto.VideoJobResponse, 0, len(jobs))
	for _, j := range jobs {
		item := &dto.VideoJobResponse{
			ID:           j.ID,
			UserID:       j.UserID,
			OriginalName: j.OriginalName,
			Status:       string(j.Status),
			FrameCount:   j.FrameCount,
			ErrorMessage: j.ErrorMessage,
			CreatedAt:    j.CreatedAt,
			UpdatedAt:    j.UpdatedAt,
			CompletedAt:  j.CompletedAt,
		}
		if j.Status == domain.StatusCompleted {
			item.DownloadURL = fmt.Sprintf("/api/v1/videos/%s/download", j.ID)
		}
		resp = append(resp, item)
	}

	c.JSON(http.StatusOK, dto.VideoListResponse{
		Total:  len(resp),
		Videos: resp,
	})
}

func (h *VideoHandler) GetByID(c *gin.Context) {
	userID, _ := c.Get(middleware.CtxUserIDKey)
	videoID := c.Param("id")

	job, err := h.videoUC.GetVideoByID(c.Request.Context(), userID.(string), videoID)
	if err != nil {
		if errors.Is(err, domain.ErrVideoNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrVideoForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao consultar vídeo: " + err.Error()})
		return
	}

	resp := &dto.VideoJobResponse{
		ID:           job.ID,
		UserID:       job.UserID,
		OriginalName: job.OriginalName,
		Status:       string(job.Status),
		FrameCount:   job.FrameCount,
		ErrorMessage: job.ErrorMessage,
		CreatedAt:    job.CreatedAt,
		UpdatedAt:    job.UpdatedAt,
		CompletedAt:  job.CompletedAt,
	}
	if job.Status == domain.StatusCompleted {
		resp.DownloadURL = fmt.Sprintf("/api/v1/videos/%s/download", job.ID)
	}

	c.JSON(http.StatusOK, resp)
}

func (h *VideoHandler) Download(c *gin.Context) {
	userID, _ := c.Get(middleware.CtxUserIDKey)
	videoID := c.Param("id")

	reader, filename, err := h.videoUC.GetVideoDownload(c.Request.Context(), userID.(string), videoID)
	if err != nil {
		if errors.Is(err, domain.ErrVideoNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrVideoForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrVideoNotReady) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao obter download: " + err.Error()})
		return
	}
	defer reader.Close()

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "application/zip")

	_, _ = io.Copy(c.Writer, reader)
}
