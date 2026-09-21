package port

import (
	"context"
	"video-processor/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type VideoJobRepository interface {
	Create(ctx context.Context, job *domain.VideoJob) error
	GetByID(ctx context.Context, id string) (*domain.VideoJob, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.VideoJob, error)
	Update(ctx context.Context, job *domain.VideoJob) error
}
