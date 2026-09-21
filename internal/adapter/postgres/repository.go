package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"video-processor/internal/domain"
	"video-processor/internal/port"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("falha ao inserir usuário: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar usuário por ID: %w", err)
	}
	return user, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`
	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar usuário por e-mail: %w", err)
	}
	return user, nil
}

type PostgresVideoJobRepository struct {
	db *sql.DB
}

func NewPostgresVideoJobRepository(db *sql.DB) *PostgresVideoJobRepository {
	return &PostgresVideoJobRepository{db: db}
}

func (r *PostgresVideoJobRepository) Create(ctx context.Context, job *domain.VideoJob) error {
	query := `
		INSERT INTO video_jobs (
			id, user_id, original_name, video_key, zip_key,
			status, frame_count, error_message, created_at, updated_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.ExecContext(ctx, query,
		job.ID,
		job.UserID,
		job.OriginalName,
		job.VideoKey,
		nullString(job.ZipKey),
		string(job.Status),
		job.FrameCount,
		nullString(job.ErrorMessage),
		job.CreatedAt,
		job.UpdatedAt,
		job.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("falha ao inserir job de vídeo: %w", err)
	}
	return nil
}

func (r *PostgresVideoJobRepository) GetByID(ctx context.Context, id string) (*domain.VideoJob, error) {
	query := `
		SELECT id, user_id, original_name, video_key, COALESCE(zip_key, ''),
		       status, frame_count, COALESCE(error_message, ''), created_at, updated_at, completed_at
		FROM video_jobs
		WHERE id = $1
	`
	job := &domain.VideoJob{}
	var statusStr string
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.UserID,
		&job.OriginalName,
		&job.VideoKey,
		&job.ZipKey,
		&statusStr,
		&job.FrameCount,
		&job.ErrorMessage,
		&job.CreatedAt,
		&job.UpdatedAt,
		&completedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrVideoNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar job por ID: %w", err)
	}

	job.Status = domain.VideoStatus(statusStr)
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return job, nil
}

func (r *PostgresVideoJobRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.VideoJob, error) {
	query := `
		SELECT id, user_id, original_name, video_key, COALESCE(zip_key, ''),
		       status, frame_count, COALESCE(error_message, ''), created_at, updated_at, completed_at
		FROM video_jobs
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar jobs por usuário: %w", err)
	}
	defer rows.Close()

	var jobs []*domain.VideoJob
	for rows.Next() {
		job := &domain.VideoJob{}
		var statusStr string
		var completedAt sql.NullTime

		if err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.OriginalName,
			&job.VideoKey,
			&job.ZipKey,
			&statusStr,
			&job.FrameCount,
			&job.ErrorMessage,
			&job.CreatedAt,
			&job.UpdatedAt,
			&completedAt,
		); err != nil {
			return nil, err
		}

		job.Status = domain.VideoStatus(statusStr)
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (r *PostgresVideoJobRepository) Update(ctx context.Context, job *domain.VideoJob) error {
	query := `
		UPDATE video_jobs
		SET zip_key = $1,
		    status = $2,
		    frame_count = $3,
		    error_message = $4,
		    updated_at = $5,
		    completed_at = $6
		WHERE id = $7
	`
	res, err := r.db.ExecContext(ctx, query,
		nullString(job.ZipKey),
		string(job.Status),
		job.FrameCount,
		nullString(job.ErrorMessage),
		time.Now().UTC(),
		job.CompletedAt,
		job.ID,
	)
	if err != nil {
		return fmt.Errorf("falha ao atualizar job de vídeo: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrVideoNotFound
	}

	return nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

var _ port.UserRepository = (*PostgresUserRepository)(nil)
var _ port.VideoJobRepository = (*PostgresVideoJobRepository)(nil)
