package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"video-processor/internal/application"
	"video-processor/internal/domain"
	"video-processor/internal/httpapi/dto"
)

type AuthHandler struct {
	authUC *application.AuthUseCase
}

func NewAuthHandler(authUC *application.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados de cadastro inválidos: " + err.Error()})
		return
	}

	out, err := h.authUC.Register(c.Request.Context(), application.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrInvalidEmail) || errors.Is(err, domain.ErrPasswordTooShort) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao registrar usuário"})
		return
	}

	c.JSON(http.StatusCreated, dto.AuthResponse{
		Token: out.Token,
		User: dto.UserResponse{
			ID:        out.User.ID,
			Name:      out.User.Name,
			Email:     out.User.Email,
			CreatedAt: out.User.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados de login inválidos: " + err.Error()})
		return
	}

	out, err := h.authUC.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha na autenticação"})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		Token: out.Token,
		User: dto.UserResponse{
			ID:        out.User.ID,
			Name:      out.User.Name,
			Email:     out.User.Email,
			CreatedAt: out.User.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}
