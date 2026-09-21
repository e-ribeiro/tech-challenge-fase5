package postgres_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"video-processor/internal/adapter/postgres"
	"video-processor/internal/domain"
)

func TestPostgresUserRepository_Create_And_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := postgres.NewPostgresUserRepository(db)
	ctx := context.Background()

	user, _ := domain.NewUser("u1", "Nome", "email@fiap.com", "senha123")

	// 1. Create
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users")).
		WithArgs(user.ID, user.Name, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(ctx, user)
	require.NoError(t, err)

	// 2. GetByID
	rows := sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "created_at", "updated_at"}).
		AddRow(user.ID, user.Name, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id = $1")).
		WithArgs("u1").
		WillReturnRows(rows)

	found, err := repo.GetByID(ctx, "u1")
	require.NoError(t, err)
	assert.Equal(t, "Nome", found.Name)

	// 3. GetByEmail
	rowsEmail := sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "created_at", "updated_at"}).
		AddRow(user.ID, user.Name, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE LOWER(email) = LOWER($1)")).
		WithArgs("email@fiap.com").
		WillReturnRows(rowsEmail)

	foundEmail, err := repo.GetByEmail(ctx, "email@fiap.com")
	require.NoError(t, err)
	assert.Equal(t, "u1", foundEmail.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresVideoJobRepository_Create_Get_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := postgres.NewPostgresVideoJobRepository(db)
	ctx := context.Background()

	job, _ := domain.NewVideoJob("j1", "u1", "video.mp4", "videos/v.mp4")

	// 1. Create
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO video_jobs")).
		WithArgs(job.ID, job.UserID, job.OriginalName, job.VideoKey, sqlmock.AnyArg(), string(job.Status), job.FrameCount, sqlmock.AnyArg(), job.CreatedAt, job.UpdatedAt, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(ctx, job)
	require.NoError(t, err)

	// 2. GetByID
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "original_name", "video_key", "zip_key", "status", "frame_count", "error_message", "created_at", "updated_at", "completed_at"}).
		AddRow("j1", "u1", "video.mp4", "videos/v.mp4", "outputs/f.zip", "COMPLETED", 15, "", now, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, original_name, video_key, COALESCE(zip_key, ''),")).
		WithArgs("j1").
		WillReturnRows(rows)

	found, err := repo.GetByID(ctx, "j1")
	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", string(found.Status))
	assert.Equal(t, 15, found.FrameCount)

	// 3. GetByUserID
	rowsUser := sqlmock.NewRows([]string{"id", "user_id", "original_name", "video_key", "zip_key", "status", "frame_count", "error_message", "created_at", "updated_at", "completed_at"}).
		AddRow("j1", "u1", "video.mp4", "videos/v.mp4", "outputs/f.zip", "COMPLETED", 15, "", now, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, original_name, video_key, COALESCE(zip_key, ''),")).
		WithArgs("u1").
		WillReturnRows(rowsUser)

	list, err := repo.GetByUserID(ctx, "u1")
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// 4. Update
	mock.ExpectExec(regexp.QuoteMeta("UPDATE video_jobs SET")).
		WithArgs(sqlmock.AnyArg(), string(job.Status), job.FrameCount, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "j1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(ctx, job)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
