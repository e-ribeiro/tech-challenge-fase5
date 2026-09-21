package storage_test

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"video-processor/internal/adapter/storage"
)

func TestLocalStorage_CRUD(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "local_storage_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	ls, err := storage.NewLocalStorage(tempDir)
	require.NoError(t, err)

	ctx := context.Background()
	key := "videos/sample.mp4"
	content := "bytes do video teste"

	// Upload
	err = ls.Upload(ctx, key, strings.NewReader(content), "video/mp4")
	require.NoError(t, err)

	// Download
	reader, err := ls.Download(ctx, key)
	require.NoError(t, err)
	defer reader.Close()

	readBytes, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, content, string(readBytes))

	// GetFileURL
	url, err := ls.GetFileURL(ctx, key)
	require.NoError(t, err)
	assert.Contains(t, url, key)

	// Delete
	err = ls.Delete(ctx, key)
	require.NoError(t, err)

	_, err = ls.Download(ctx, key)
	assert.Error(t, err)
}
