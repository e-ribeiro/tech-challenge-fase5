package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"video-processor/internal/application"
)

const (
	CtxUserIDKey    = "user_id"
	CtxUserEmailKey = "user_email"
	CtxUserNameKey  = "user_name"
)

func AuthMiddleware(authUC *application.AuthUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := ""

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenStr = strings.TrimSpace(parts[1])
			}
		}

		// Fallback para query parameter (útil para download direto no navegador via GET)
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "token de autorização não fornecido",
			})
			return
		}

		claims, err := authUC.ValidateToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "token inválido ou expirado",
			})
			return
		}

		c.Set(CtxUserIDKey, claims.UserID)
		c.Set(CtxUserEmailKey, claims.Email)
		c.Set(CtxUserNameKey, claims.Name)

		c.Next()
	}
}
