package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"video-processor/internal/adapter/memory"
	"video-processor/internal/application"
	"video-processor/internal/domain"
)

func TestAuthUseCase_Register_Success(t *testing.T) {
	repo := memory.NewMemoryUserRepository()
	uc := application.NewAuthUseCase(repo, "secret123", 1*time.Hour)

	ctx := context.Background()
	out, err := uc.Register(ctx, application.RegisterInput{
		Name:     "Investidor FIAP",
		Email:    "investidor@fiap.com",
		Password: "senhaSegura123",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, out.Token)
	assert.Equal(t, "Investidor FIAP", out.User.Name)
	assert.Equal(t, "investidor@fiap.com", out.User.Email)

	// Valida se o token gerado é válido
	claims, err := uc.ValidateToken(out.Token)
	require.NoError(t, err)
	assert.Equal(t, out.User.ID, claims.UserID)
	assert.Equal(t, out.User.Email, claims.Email)
}

func TestAuthUseCase_Register_DuplicateEmail(t *testing.T) {
	repo := memory.NewMemoryUserRepository()
	uc := application.NewAuthUseCase(repo, "secret123", 1*time.Hour)

	ctx := context.Background()
	_, err := uc.Register(ctx, application.RegisterInput{
		Name:     "Primeiro",
		Email:    "duplicado@fiap.com",
		Password: "senhaSegura123",
	})
	require.NoError(t, err)

	_, err = uc.Register(ctx, application.RegisterInput{
		Name:     "Segundo",
		Email:    "duplicado@fiap.com",
		Password: "outraSenha123",
	})
	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
}

func TestAuthUseCase_Login_Success(t *testing.T) {
	repo := memory.NewMemoryUserRepository()
	uc := application.NewAuthUseCase(repo, "secret123", 1*time.Hour)

	ctx := context.Background()
	_, err := uc.Register(ctx, application.RegisterInput{
		Name:     "Carlos",
		Email:    "carlos@fiap.com",
		Password: "senhaValida123",
	})
	require.NoError(t, err)

	out, err := uc.Login(ctx, "carlos@fiap.com", "senhaValida123")
	require.NoError(t, err)
	assert.NotEmpty(t, out.Token)
	assert.Equal(t, "carlos@fiap.com", out.User.Email)
}

func TestAuthUseCase_Login_InvalidCredentials(t *testing.T) {
	repo := memory.NewMemoryUserRepository()
	uc := application.NewAuthUseCase(repo, "secret123", 1*time.Hour)

	ctx := context.Background()
	_, err := uc.Register(ctx, application.RegisterInput{
		Name:     "Carlos",
		Email:    "carlos@fiap.com",
		Password: "senhaValida123",
	})
	require.NoError(t, err)

	// Senha incorreta
	_, err = uc.Login(ctx, "carlos@fiap.com", "senhaIncorreta")
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)

	// Usuário inexistente
	_, err = uc.Login(ctx, "naoexiste@fiap.com", "senhaValida123")
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestAuthUseCase_ValidateToken_Invalid(t *testing.T) {
	repo := memory.NewMemoryUserRepository()
	uc := application.NewAuthUseCase(repo, "secret123", 1*time.Hour)

	claims, err := uc.ValidateToken("token-invalido-xyz")
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
	assert.Nil(t, claims)
}
