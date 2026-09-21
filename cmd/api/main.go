package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"video-processor/internal/adapter/memory"
	"video-processor/internal/adapter/postgres"
	"video-processor/internal/adapter/rabbitmq"
	"video-processor/internal/adapter/storage"
	"video-processor/internal/application"
	"video-processor/internal/httpapi"
	"video-processor/internal/port"
)

func main() {
	portStr := getEnv("PORT", "8080")
	jwtSecret := getEnv("JWT_SECRET", "super-secret-jwt-key-fiapx-hackathon-2026")
	storageDriver := getEnv("STORAGE_DRIVER", "postgres") // "postgres" ou "memory"

	var userRepo port.UserRepository
	var videoRepo port.VideoJobRepository
	var objStorage port.Storage
	var queueProducer port.QueueProducer

	if storageDriver == "memory" {
		log.Println("⚡ Modo STORAGE_DRIVER=memory ativado (desenvolvimento em memória)")
		memUserRepo := memory.NewMemoryUserRepository()
		memVideoRepo := memory.NewMemoryVideoJobRepository()
		memStorage := memory.NewMemoryStorage()
		memQueue := memory.NewMemoryQueue(100)

		userRepo = memUserRepo
		videoRepo = memVideoRepo
		objStorage = memStorage
		queueProducer = memQueue
	} else {
		// PostgreSQL
		dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/fiapx_videos?sslmode=disable")
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatalf("Falha ao abrir conexão com PostgreSQL: %v", err)
		}
		defer db.Close()

		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(10)
		db.SetConnMaxLifetime(5 * time.Minute)

		ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
		if err := db.PingContext(ctxPing); err != nil {
			log.Printf("⚠️ Aviso: Não foi possível conectar ao PostgreSQL: %v. Usando fallback em memória.", err)
			memUserRepo := memory.NewMemoryUserRepository()
			memVideoRepo := memory.NewMemoryVideoJobRepository()
			userRepo = memUserRepo
			videoRepo = memVideoRepo
		} else {
			log.Println("✅ Conectado ao PostgreSQL com sucesso!")
			userRepo = postgres.NewPostgresUserRepository(db)
			videoRepo = postgres.NewPostgresVideoJobRepository(db)
		}
		cancelPing()

		// MinIO / S3 Storage
		s3Endpoint := getEnv("S3_ENDPOINT", "localhost:9000")
		s3AccessKey := getEnv("S3_ACCESS_KEY", "minioadmin")
		s3SecretKey := getEnv("S3_SECRET_KEY", "minioadmin")
		s3Bucket := getEnv("S3_BUCKET", "fiapx-videos")
		s3UseSSL, _ := strconv.ParseBool(getEnv("S3_USE_SSL", "false"))

		s3Client, err := storage.NewS3Storage(s3Endpoint, s3AccessKey, s3SecretKey, s3Bucket, s3UseSSL)
		if err != nil {
			log.Printf("⚠️ Aviso: Falha ao conectar ao MinIO/S3 (%v). Usando armazenamento local em 'uploads'.", err)
			localStorage, _ := storage.NewLocalStorage("uploads")
			objStorage = localStorage
		} else {
			log.Println("✅ Conectado ao Object Storage (S3/MinIO) com sucesso!")
			objStorage = s3Client
		}

		// RabbitMQ
		rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
		broker, err := rabbitmq.NewRabbitMQBroker(rabbitURL)
		if err != nil {
			log.Printf("⚠️ Aviso: Falha ao conectar no RabbitMQ (%v). Usando mensageria em memória.", err)
			memQueue := memory.NewMemoryQueue(100)
			queueProducer = memQueue
		} else {
			log.Println("✅ Conectado ao RabbitMQ com sucesso!")
			defer broker.Close()
			queueProducer = broker
		}
	}

	// Use Cases
	authUC := application.NewAuthUseCase(userRepo, jwtSecret, 24*time.Hour)
	videoUC := application.NewVideoUseCase(videoRepo, objStorage, queueProducer)

	router := httpapi.SetupRouter(httpapi.ServerConfig{
		AuthUseCase:  authUC,
		VideoUseCase: videoUC,
	})

	srv := &http.Server{
		Addr:    ":" + portStr,
		Handler: router,
	}

	go func() {
		log.Printf("🚀 FIAP X Video API rodando na porta %s", portStr)
		log.Printf("🌐 Acesse a interface web em: http://localhost:%s", portStr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro no servidor HTTP: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Encerrando servidor da API com segurança...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Erro durante o encerramento forçado: %v", err)
	}

	log.Println("👋 Servidor encerrado.")
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}
