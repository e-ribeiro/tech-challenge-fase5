package memory

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"video-processor/internal/port"
)

type MemoryStorage struct {
	mu    sync.RWMutex
	files map[string][]byte
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		files: make(map[string][]byte),
	}
}

func (s *MemoryStorage) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	s.files[key] = data
	return nil
}

func (s *MemoryStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.files[key]
	if !ok {
		return nil, fmt.Errorf("arquivo %s não encontrado no storage", key)
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *MemoryStorage) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.files, key)
	return nil
}

func (s *MemoryStorage) GetFileURL(ctx context.Context, key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.files[key]; !ok {
		return "", fmt.Errorf("arquivo %s não encontrado", key)
	}
	return "/files/" + key, nil
}

var _ port.Storage = (*MemoryStorage)(nil)
