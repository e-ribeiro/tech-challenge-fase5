package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"video-processor/internal/adapter/mailer"
	"video-processor/internal/adapter/postgres"
	"video-processor/internal/adapter/processor"
	"video-processor/internal/adapter/rabbitmq"
	"video-processor/internal/adapter/storage"
	"video-processor/internal/application"
	"video-processor/internal/port"
)

func main() {
	log.Println("⚙️ Iniciando FIAP X Video Worker...")

	// 1. PostgreSQL
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/fiapx_videos?sslmode=disable")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Falha ao abrir conexão com PostgreSQL: %v", err)
	}
	defer db.Close()

	ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	if err := db.PingContext(ctxPing); err != nil {
		log.Fatalf("Falha ao conectar ao PostgreSQL: %v", err)
	}
	cancelPing()
	log.Println("✅ Worker conectado ao PostgreSQL.")

	videoRepo := postgres.NewPostgresVideoJobRepository(db)

	// 2. Storage MinIO / S3
	s3Endpoint := getEnv("S3_ENDPOINT", "localhost:9000")
	s3AccessKey := getEnv("S3_ACCESS_KEY", "minioadmin")
	s3SecretKey := getEnv("S3_SECRET_KEY", "minioadmin")
	s3Bucket := getEnv("S3_BUCKET", "fiapx-videos")
	s3UseSSL, _ := strconv.ParseBool(getEnv("S3_USE_SSL", "false"))

	s3Client, err := storage.NewS3Storage(s3Endpoint, s3AccessKey, s3SecretKey, s3Bucket, s3UseSSL)
	if err != nil {
		log.Printf("⚠️ Falha ao conectar ao MinIO/S3 (%v). Usando armazenamento local.", err)
	}
	var objStorage port.Storage
	if s3Client != nil {
		objStorage = s3Client
	} else {
		localStorage, _ := storage.NewLocalStorage("uploads")
		objStorage = localStorage
	}

	// 3. RabbitMQ
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	broker, err := rabbitmq.NewRabbitMQBroker(rabbitURL)
	if err != nil {
		log.Fatalf("Falha ao conectar no RabbitMQ: %v", err)
	}
	defer broker.Close()
	log.Println("✅ Worker conectado ao RabbitMQ.")

	// 4. Video Processor (FFmpeg)
	videoProc := processor.NewFFmpegVideoProcessor()

	// 5. Worker Use Case
	workerUC := application.NewWorkerUseCase(videoRepo, objStorage, videoProc, broker)

	// 6. Notifier & Notification Worker
	smtpHost := getEnv("SMTP_HOST", "localhost")
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "1025"))
	smtpUser := getEnv("SMTP_USER", "")
	smtpPass := getEnv("SMTP_PASS", "")
	smtpFrom := getEnv("SMTP_FROM", "no-reply@fiapx.com")

	smtpMailer := mailer.NewSMTPMailer(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom)
	notificationUC := application.NewNotificationUseCase(smtpMailer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Inicia consumidor de processamento de vídeos
	go func() {
		log.Println("🎬 Aguardando vídeos na fila 'video.process.queue'...")
		err := broker.ConsumeVideoProcess(ctx, func(ctx context.Context, msg *port.VideoProcessMessage) error {
			log.Printf("📥 [Worker] Mensagem recebida! Job ID: %s | Vídeo: %s", msg.JobID, msg.OriginalName)
			return workerUC.ProcessJob(ctx, msg)
		})
		if err != nil && ctx.Err() == nil {
			log.Printf("Erro no consumidor de vídeos: %v", err)
		}
	}()

	// Inicia consumidor de notificações de erro/sucesso
	go func() {
		log.Println("📬 Aguardando notificações na fila 'video.notification.queue'...")
		err := broker.ConsumeNotification(ctx, func(ctx context.Context, msg *port.NotificationMessage) error {
			log.Printf("📧 [Notificação] Enviando e-mail de %s para %s (Job %s)", msg.Type, msg.UserEmail, msg.JobID)
			return notificationUC.HandleNotification(ctx, msg)
		})
		if err != nil && ctx.Err() == nil {
			log.Printf("Erro no consumidor de notificações: %v", err)
		}
	}()

	// Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("🛑 Encerrando Worker com segurança...")
	cancel()
	time.Sleep(1 * time.Second)
	log.Println("👋 Worker finalizado.")
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}
