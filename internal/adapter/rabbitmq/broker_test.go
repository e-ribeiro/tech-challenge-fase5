package rabbitmq_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"video-processor/internal/adapter/rabbitmq"
	"video-processor/internal/port"
)

func TestRabbitMQ_ConstantsAndMessageSerialization(t *testing.T) {
	assert.Equal(t, "video.process.queue", rabbitmq.VideoProcessQueue)
	assert.Equal(t, "video.process.dlq", rabbitmq.VideoProcessDLQ)
	assert.Equal(t, "video.notification.queue", rabbitmq.NotificationQueue)

	msg := &port.VideoProcessMessage{
		JobID:        "job-1",
		UserID:       "u-1",
		UserEmail:    "test@fiap.com",
		OriginalName: "video.mp4",
		VideoKey:     "videos/v.mp4",
		CreatedAt:    time.Now().UTC(),
	}

	bytes, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded port.VideoProcessMessage
	err = json.Unmarshal(bytes, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "job-1", decoded.JobID)
	assert.Equal(t, "test@fiap.com", decoded.UserEmail)
}
