package port

import (
	"context"
)

type ProcessResult struct {
	ZipLocalPath string
	FrameCount   int
}

type VideoProcessor interface {
	ProcessVideoToZip(ctx context.Context, localVideoPath string, fps int) (*ProcessResult, error)
}
