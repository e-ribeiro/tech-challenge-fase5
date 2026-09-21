package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"video-processor/internal/domain"
)

func TestNewVideoJob_Success(t *testing.T) {
	job, err := domain.NewVideoJob("job-1", "user-1", "video.mp4", "videos/video.mp4")
	require.NoError(t, err)
	assert.Equal(t, "job-1", job.ID)
	assert.Equal(t, "user-1", job.UserID)
	assert.Equal(t, "video.mp4", job.OriginalName)
	assert.Equal(t, domain.StatusPending, job.Status)
	assert.Equal(t, 0, job.FrameCount)
	assert.Nil(t, job.CompletedAt)
}

func TestNewVideoJob_InvalidExtension(t *testing.T) {
	job, err := domain.NewVideoJob("job-1", "user-1", "documento.pdf", "videos/documento.pdf")
	assert.ErrorIs(t, err, domain.ErrUnsupportedVideoType)
	assert.Nil(t, job)
}

func TestVideoJob_StateTransitions(t *testing.T) {
	job, err := domain.NewVideoJob("job-1", "user-1", "video.avi", "videos/video.avi")
	require.NoError(t, err)

	// Mark processing
	job.MarkProcessing()
	assert.Equal(t, domain.StatusProcessing, job.Status)

	// Mark completed
	job.MarkCompleted("outputs/frames_job1.zip", 42)
	assert.Equal(t, domain.StatusCompleted, job.Status)
	assert.Equal(t, 42, job.FrameCount)
	assert.Equal(t, "outputs/frames_job1.zip", job.ZipKey)
	assert.NotNil(t, job.CompletedAt)
	assert.Empty(t, job.ErrorMessage)

	// Mark failed
	job.MarkFailed("erro de codec")
	assert.Equal(t, domain.StatusFailed, job.Status)
	assert.Equal(t, "erro de codec", job.ErrorMessage)
}

func TestIsValidVideoExtension(t *testing.T) {
	valid := []string{"video.mp4", "test.AVI", "clip.mov", "sample.MKV", "f.webm"}
	for _, f := range valid {
		assert.True(t, domain.IsValidVideoExtension(f), "deve ser válido: %s", f)
	}

	invalid := []string{"file.exe", "photo.jpg", "text.txt", "archive.zip"}
	for _, f := range invalid {
		assert.False(t, domain.IsValidVideoExtension(f), "deve ser inválido: %s", f)
	}
}
