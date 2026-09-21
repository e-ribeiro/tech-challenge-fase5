package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"video-processor/internal/application"
	"video-processor/internal/httpapi/handler"
	"video-processor/internal/httpapi/middleware"
)

type ServerConfig struct {
	AuthUseCase  *application.AuthUseCase
	VideoUseCase *application.VideoUseCase
}

func SetupRouter(cfg ServerConfig) *gin.Engine {
	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Interface Web integrada para os investidores e avaliação da banca
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, GetAppHTML())
	})

	// Health Checks para Kubernetes
	healthHandler := handler.NewHealthHandler()
	r.GET("/health/live", healthHandler.Live)
	r.GET("/health/ready", healthHandler.Ready)

	// Handlers de Domínio
	authHandler := handler.NewAuthHandler(cfg.AuthUseCase)
	videoHandler := handler.NewVideoHandler(cfg.VideoUseCase)

	// Rotas da API v1
	v1 := r.Group("/api/v1")
	{
		// Rotas de Autenticação
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Rotas Protegidas de Vídeo
		videos := v1.Group("/videos")
		videos.Use(middleware.AuthMiddleware(cfg.AuthUseCase))
		{
			videos.POST("/upload", videoHandler.Upload)
			videos.GET("", videoHandler.List)
			videos.GET("/:id", videoHandler.GetByID)
			videos.GET("/:id/download", videoHandler.Download)
		}
	}

	return r
}
