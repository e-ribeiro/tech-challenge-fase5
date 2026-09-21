package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"video-processor/internal/port"
)

type LocalStorage struct {
	baseDir string
}

func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("falha ao criar diretório base de storage: %w", err)
	}
	return &LocalStorage{baseDir: baseDir}, nil
}

func (s *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	fullPath := filepath.Join(s.baseDir, key)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("falha ao criar arquivo local: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	return err
}

func (s *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.baseDir, key)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir arquivo local: %w", err)
	}
	return file, nil
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(s.baseDir, key)
	return os.Remove(fullPath)
}

func (s *LocalStorage) GetFileURL(ctx context.Context, key string) (string, error) {
	return "/files/" + key, nil
}

var _ port.Storage = (*LocalStorage)(nil)
