package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"video-processor/internal/adapter/memory"
	"video-processor/internal/application"
	"video-processor/internal/domain"
	"video-processor/internal/httpapi/middleware"
)

func TestAuthMiddleware_Scenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := memory.NewMemoryUserRepository()
	authUC := application.NewAuthUseCase(repo, "secret123", 1*time.Hour)

	user, err := domain.NewUser("u1", "Nome", "email@fiap.com", "senha123")
	require.NoError(t, err)
	_ = repo.Create(t.Context(), user)

	validToken, err := authUC.GenerateToken(user)
	require.NoError(t, err)

	r := gin.New()
	r.Use(middleware.AuthMiddleware(authUC))
	r.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get(middleware.CtxUserIDKey)
		userEmail, _ := c.Get(middleware.CtxUserEmailKey)
		userName, _ := c.Get(middleware.CtxUserNameKey)
		c.JSON(http.StatusOK, gin.H{
			"user_id":    userID,
			"user_email": userEmail,
			"user_name":  userName,
		})
	})

	// 1. Sem token
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusUnauthorized, w1.Code)

	// 2. Token inválido
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req2.Header.Set("Authorization", "Bearer invalid-token")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)

	// 3. Bearer Token válido no Header
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req3.Header.Set("Authorization", "Bearer "+validToken)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
	assert.Contains(t, w3.Body.String(), "email@fiap.com")

	// 4. Token válido via Query Param (?token=...)
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodGet, "/protected?token="+validToken, nil)
	r.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)
	assert.Contains(t, w4.Body.String(), "email@fiap.com")
}
