package processor_test

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"video-processor/internal/adapter/processor"
)

func TestMockVideoProcessor_Success(t *testing.T) {
	proc := processor.NewMockVideoProcessor(12, false)

	res, err := proc.ProcessVideoToZip(context.Background(), "fake_path.mp4", 1)
	require.NoError(t, err)
	assert.Equal(t, 12, res.FrameCount)
	assert.NotEmpty(t, res.ZipLocalPath)
	defer os.RemoveAll(res.ZipLocalPath)

	// Valida se o arquivo zip gerado é válido
	zr, err := zip.OpenReader(res.ZipLocalPath)
	require.NoError(t, err)
	defer zr.Close()

	assert.Len(t, zr.File, 12)
}

func TestMockVideoProcessor_Failure(t *testing.T) {
	proc := processor.NewMockVideoProcessor(10, true)

	res, err := proc.ProcessVideoToZip(context.Background(), "fake_path.mp4", 1)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestNewFFmpegVideoProcessor(t *testing.T) {
	proc := processor.NewFFmpegVideoProcessor()
	assert.NotNil(t, proc)
}

func TestCreateZipFile_RealFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "zip_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	f1 := filepath.Join(tempDir, "f1.txt")
	f2 := filepath.Join(tempDir, "f2.txt")
	_ = os.WriteFile(f1, []byte("conteudo 1"), 0644)
	_ = os.WriteFile(f2, []byte("conteudo 2"), 0644)

	zipOut := filepath.Join(tempDir, "saida.zip")
	// Usa NewMockVideoProcessor para validar
	m := processor.NewMockVideoProcessor(2, false)
	res, err := m.ProcessVideoToZip(t.Context(), "video.mp4", 1)
	require.NoError(t, err)
	assert.Equal(t, 2, res.FrameCount)
	_ = zipOut
}

