package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"video-processor/internal/domain"
)

func TestNewUser_Success(t *testing.T) {
	user, err := domain.NewUser("uuid-1", "João Silva", "joao@exemplo.com", "senha123")
	require.NoError(t, err)
	assert.Equal(t, "uuid-1", user.ID)
	assert.Equal(t, "João Silva", user.Name)
	assert.Equal(t, "joao@exemplo.com", user.Email)
	assert.NotEmpty(t, user.PasswordHash)
	assert.True(t, user.CheckPassword("senha123"))
	assert.False(t, user.CheckPassword("senhaErrada"))
}

func TestNewUser_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		userName    string
		email       string
		password    string
		expectedErr error
	}{
		{
			name:        "email vazio",
			id:          "1",
			userName:    "Teste",
			email:       "",
			password:    "senha123",
			expectedErr: domain.ErrInvalidEmail,
		},
		{
			name:        "email sem arroba",
			id:          "1",
			userName:    "Teste",
			email:       "invalido.com",
			password:    "senha123",
			expectedErr: domain.ErrInvalidEmail,
		},
		{
			name:        "senha curta",
			id:          "1",
			userName:    "Teste",
			email:       "teste@fiap.com",
			password:    "12345",
			expectedErr: domain.ErrPasswordTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := domain.NewUser(tt.id, tt.userName, tt.email, tt.password)
			assert.Nil(t, u)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}
