package memory

import (
	"context"
	"sync"

	"video-processor/internal/domain"
	"video-processor/internal/port"
)

type MemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (r *MemoryUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if u.Email == user.Email {
			return domain.ErrUserAlreadyExists
		}
	}

	copyUser := *user
	r.users[user.ID] = &copyUser
	return nil
}

func (r *MemoryUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	copyUser := *user
	return &copyUser, nil
}

func (r *MemoryUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			copyUser := *user
			return &copyUser, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

type MemoryVideoJobRepository struct {
	mu   sync.RWMutex
	jobs map[string]*domain.VideoJob
}

func NewMemoryVideoJobRepository() *MemoryVideoJobRepository {
	return &MemoryVideoJobRepository{
		jobs: make(map[string]*domain.VideoJob),
	}
}

func (r *MemoryVideoJobRepository) Create(ctx context.Context, job *domain.VideoJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copyJob := *job
	r.jobs[job.ID] = &copyJob
	return nil
}

func (r *MemoryVideoJobRepository) GetByID(ctx context.Context, id string) (*domain.VideoJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	job, ok := r.jobs[id]
	if !ok {
		return nil, domain.ErrVideoNotFound
	}
	copyJob := *job
	return &copyJob, nil
}

func (r *MemoryVideoJobRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.VideoJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.VideoJob
	for _, job := range r.jobs {
		if job.UserID == userID {
			copyJob := *job
			result = append(result, &copyJob)
		}
	}
	return result, nil
}

func (r *MemoryVideoJobRepository) Update(ctx context.Context, job *domain.VideoJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.jobs[job.ID]; !ok {
		return domain.ErrVideoNotFound
	}

	copyJob := *job
	r.jobs[job.ID] = &copyJob
	return nil
}

var _ port.UserRepository = (*MemoryUserRepository)(nil)
var _ port.VideoJobRepository = (*MemoryVideoJobRepository)(nil)
