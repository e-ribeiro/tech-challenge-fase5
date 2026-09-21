package domain

import (
	"path/filepath"
	"strings"
	"time"
)

type VideoStatus string

const (
	StatusPending    VideoStatus = "PENDING"
	StatusProcessing VideoStatus = "PROCESSING"
	StatusCompleted  VideoStatus = "COMPLETED"
	StatusFailed     VideoStatus = "FAILED"
)

var allowedVideoExtensions = map[string]bool{
	".mp4":  true,
	".avi":  true,
	".mov":  true,
	".mkv":  true,
	".wmv":  true,
	".flv":  true,
	".webm": true,
}

func IsValidVideoExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return allowedVideoExtensions[ext]
}

type VideoJob struct {
	ID           string      `json:"id"`
	UserID       string      `json:"user_id"`
	OriginalName string      `json:"original_name"`
	VideoKey     string      `json:"video_key"`
	ZipKey       string      `json:"zip_key,omitempty"`
	Status       VideoStatus `json:"status"`
	FrameCount   int         `json:"frame_count"`
	ErrorMessage string      `json:"error_message,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	CompletedAt  *time.Time  `json:"completed_at,omitempty"`
}

func NewVideoJob(id, userID, originalName, videoKey string) (*VideoJob, error) {
	if !IsValidVideoExtension(originalName) {
		return nil, ErrUnsupportedVideoType
	}

	now := time.Now().UTC()
	return &VideoJob{
		ID:           id,
		UserID:       userID,
		OriginalName: originalName,
		VideoKey:     videoKey,
		Status:       StatusPending,
		FrameCount:   0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (v *VideoJob) MarkProcessing() {
	v.Status = StatusProcessing
	v.UpdatedAt = time.Now().UTC()
}

func (v *VideoJob) MarkCompleted(zipKey string, frameCount int) {
	now := time.Now().UTC()
	v.Status = StatusCompleted
	v.ZipKey = zipKey
	v.FrameCount = frameCount
	v.UpdatedAt = now
	v.CompletedAt = &now
	v.ErrorMessage = ""
}

func (v *VideoJob) MarkFailed(errMsg string) {
	now := time.Now().UTC()
	v.Status = StatusFailed
	v.ErrorMessage = errMsg
	v.UpdatedAt = now
}
