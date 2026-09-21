package processor

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"video-processor/internal/port"
)

type FFmpegVideoProcessor struct{}

func NewFFmpegVideoProcessor() *FFmpegVideoProcessor {
	return &FFmpegVideoProcessor{}
}

func (p *FFmpegVideoProcessor) ProcessVideoToZip(ctx context.Context, localVideoPath string, fps int) (*port.ProcessResult, error) {
	if fps <= 0 {
		fps = 1
	}

	tempDir, err := os.MkdirTemp("", "ffmpeg_frames_*")
	if err != nil {
		return nil, fmt.Errorf("falha ao criar pasta de frames: %w", err)
	}
	defer os.RemoveAll(tempDir)

	framePattern := filepath.Join(tempDir, "frame_%04d.png")

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", localVideoPath,
		"-vf", fmt.Sprintf("fps=%d", fps),
		"-y",
		framePattern,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("erro na execução do ffmpeg: %w, log: %s", err, string(output))
	}

	frames, err := filepath.Glob(filepath.Join(tempDir, "*.png"))
	if err != nil || len(frames) == 0 {
		return nil, fmt.Errorf("nenhum frame foi extraído do vídeo")
	}

	// Cria o arquivo ZIP temporário
	zipDir, err := os.MkdirTemp("", "ffmpeg_zip_*")
	if err != nil {
		return nil, fmt.Errorf("falha ao criar pasta para zip: %w", err)
	}
	zipLocalPath := filepath.Join(zipDir, "output_frames.zip")

	if err := createZipFile(frames, zipLocalPath); err != nil {
		_ = os.RemoveAll(zipDir)
		return nil, fmt.Errorf("falha ao criar zip: %w", err)
	}

	return &port.ProcessResult{
		ZipLocalPath: zipLocalPath,
		FrameCount:   len(frames),
	}, nil
}

func createZipFile(files []string, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range files {
		if err := addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filePath)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// MockVideoProcessor para testes unitários sem necessidade do binário ffmpeg no host
type MockVideoProcessor struct {
	ShouldFail bool
	FrameCount int
}

func NewMockVideoProcessor(frameCount int, shouldFail bool) *MockVideoProcessor {
	if frameCount <= 0 {
		frameCount = 10
	}
	return &MockVideoProcessor{
		FrameCount: frameCount,
		ShouldFail: shouldFail,
	}
}

func (m *MockVideoProcessor) ProcessVideoToZip(ctx context.Context, localVideoPath string, fps int) (*port.ProcessResult, error) {
	if m.ShouldFail {
		return nil, fmt.Errorf("falha simulada no processamento de vídeo")
	}

	zipDir, err := os.MkdirTemp("", "mock_zip_*")
	if err != nil {
		return nil, err
	}
	zipLocalPath := filepath.Join(zipDir, "output_frames.zip")

	zipFile, err := os.Create(zipLocalPath)
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()

	zw := zip.NewWriter(zipFile)
	for i := 1; i <= m.FrameCount; i++ {
		w, _ := zw.Create(fmt.Sprintf("frame_%04d.png", i))
		_, _ = w.Write([]byte(fmt.Sprintf("conteudo_frame_%d", i)))
	}
	_ = zw.Close()

	return &port.ProcessResult{
		ZipLocalPath: zipLocalPath,
		FrameCount:   m.FrameCount,
	}, nil
}

var _ port.VideoProcessor = (*FFmpegVideoProcessor)(nil)
var _ port.VideoProcessor = (*MockVideoProcessor)(nil)
