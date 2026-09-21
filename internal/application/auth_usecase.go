package application

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"video-processor/internal/domain"
	"video-processor/internal/port"
)

type TokenClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

type AuthUseCase struct {
	userRepo  port.UserRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewAuthUseCase(userRepo port.UserRepository, jwtSecret string, tokenTTL time.Duration) *AuthUseCase {
	if tokenTTL <= 0 {
		tokenTTL = 24 * time.Hour
	}
	return &AuthUseCase{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  tokenTTL,
	}
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type AuthOutput struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

func (uc *AuthUseCase) Register(ctx context.Context, input RegisterInput) (*AuthOutput, error) {
	existing, err := uc.userRepo.GetByEmail(ctx, input.Email)
	if err == nil && existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	id := uuid.New().String()
	user, err := domain.NewUser(id, input.Name, input.Email, input.Password)
	if err != nil {
		return nil, err
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := uc.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthOutput{
		Token: token,
		User:  user,
	}, nil
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*AuthOutput, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.CheckPassword(password) {
		return nil, domain.ErrInvalidCredentials
	}

	token, err := uc.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthOutput{
		Token: token,
		User:  user,
	}, nil
}

func (uc *AuthUseCase) GenerateToken(user *domain.User) (string, error) {
	claims := TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(uc.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(uc.jwtSecret)
}

func (uc *AuthUseCase) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return uc.jwtSecret, nil
	})

	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrUnauthorized
	}

	return claims, nil
}
