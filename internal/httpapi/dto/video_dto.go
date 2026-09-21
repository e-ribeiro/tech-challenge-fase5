package dto

import "time"

type VideoJobResponse struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	OriginalName string     `json:"original_name"`
	Status       string     `json:"status"`
	FrameCount   int        `json:"frame_count"`
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	DownloadURL  string     `json:"download_url,omitempty"`
}

type VideoListResponse struct {
	Total  int                 `json:"total"`
	Videos []*VideoJobResponse `json:"videos"`
}

type UploadVideoResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
