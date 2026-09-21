package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"video-processor/internal/port"
)

type S3Storage struct {
	client *minio.Client
	bucket string
}

func NewS3Storage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*S3Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao inicializar cliente MinIO/S3: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("falha ao verificar existência do bucket %s: %w", bucket, err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("falha ao criar bucket %s: %w", bucket, err)
		}
	}

	return &S3Storage{
		client: client,
		bucket: bucket,
	}, nil
}

func (s *S3Storage) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	// Se o tamanho não for conhecido, lê para buffer ou usa PutObject com stream
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("falha ao ler stream para upload: %w", err)
	}

	buf := bytes.NewReader(data)
	_, err = s.client.PutObject(ctx, s.bucket, key, buf, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("falha ao enviar objeto para o bucket: %w", err)
	}
	return nil
}

func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("falha ao baixar objeto do bucket: %w", err)
	}
	// Verifica se o objeto existe
	_, err = object.Stat()
	if err != nil {
		_ = object.Close()
		return nil, fmt.Errorf("objeto %s não encontrado no bucket: %w", key, err)
	}
	return object, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

func (s *S3Storage) GetFileURL(ctx context.Context, key string) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, key, 2*time.Hour, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

var _ port.Storage = (*S3Storage)(nil)
