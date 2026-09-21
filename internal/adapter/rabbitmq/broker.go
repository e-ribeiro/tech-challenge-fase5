package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"video-processor/internal/port"
)

const (
	VideoProcessQueue = "video.process.queue"
	VideoProcessDLQ   = "video.process.dlq"
	VideoProcessDLX   = "video.process.dlx"
	NotificationQueue = "video.notification.queue"
	NotificationDLQ   = "video.notification.dlq"
	NotificationDLX   = "video.notification.dlx"
)

type RabbitMQBroker struct {
	url      string
	conn     *amqp.Connection
	ch       *amqp.Channel
	mu       sync.Mutex
	isClosed bool
}

func NewRabbitMQBroker(url string) (*RabbitMQBroker, error) {
	broker := &RabbitMQBroker{url: url}
	if err := broker.connect(); err != nil {
		return nil, err
	}
	return broker, nil
}

func (b *RabbitMQBroker) connect() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	var err error
	b.conn, err = amqp.Dial(b.url)
	if err != nil {
		return fmt.Errorf("falha ao conectar no RabbitMQ: %w", err)
	}

	b.ch, err = b.conn.Channel()
	if err != nil {
		_ = b.conn.Close()
		return fmt.Errorf("falha ao abrir canal RabbitMQ: %w", err)
	}

	// Configura QoS (prefetch 1 para distribuição balanceada de carga entre workers)
	if err := b.ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("falha ao configurar QoS: %w", err)
	}

	if err := b.setupQueues(); err != nil {
		return err
	}

	return nil
}

func (b *RabbitMQBroker) setupQueues() error {
	// 1. DLX e DLQ para Processamento de Vídeo
	if err := b.ch.ExchangeDeclare(VideoProcessDLX, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := b.ch.QueueDeclare(VideoProcessDLQ, true, false, false, false, nil); err != nil {
		return err
	}
	if err := b.ch.QueueBind(VideoProcessDLQ, "dlq", VideoProcessDLX, false, nil); err != nil {
		return err
	}

	// 2. Fila Principal de Processamento de Vídeo com Dead-Letter configurado
	videoArgs := amqp.Table{
		"x-dead-letter-exchange":    VideoProcessDLX,
		"x-dead-letter-routing-key": "dlq",
	}
	if _, err := b.ch.QueueDeclare(VideoProcessQueue, true, false, false, false, videoArgs); err != nil {
		return err
	}

	// 3. Fila de Notificação
	if _, err := b.ch.QueueDeclare(NotificationQueue, true, false, false, false, nil); err != nil {
		return err
	}

	return nil
}

func (b *RabbitMQBroker) PublishVideoProcess(ctx context.Context, msg *port.VideoProcessMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return b.ch.PublishWithContext(ctx,
		"",                // exchange padrão (direct para fila)
		VideoProcessQueue, // routing key = nome da fila
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

func (b *RabbitMQBroker) PublishNotification(ctx context.Context, msg *port.NotificationMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return b.ch.PublishWithContext(ctx,
		"",
		NotificationQueue,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

func (b *RabbitMQBroker) ConsumeVideoProcess(ctx context.Context, handler func(ctx context.Context, msg *port.VideoProcessMessage) error) error {
	deliveries, err := b.ch.Consume(
		VideoProcessQueue,
		"video-worker",
		false, // auto-ack = false (ACK manual obrigatório)
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("falha ao iniciar consumidor da fila de vídeo: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				return nil
			}

			var msg port.VideoProcessMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("Erro de parse na mensagem, enviando para DLQ: %v", err)
				_ = d.Nack(false, false) // envia para DLQ sem reenfileirar
				continue
			}

			if err := handler(ctx, &msg); err != nil {
				log.Printf("Falha no processamento do vídeo %s: %v", msg.JobID, err)
				_ = d.Nack(false, false) // envia para DLQ
			} else {
				_ = d.Ack(false) // confirma sucesso
			}
		}
	}
}

func (b *RabbitMQBroker) ConsumeNotification(ctx context.Context, handler func(ctx context.Context, msg *port.NotificationMessage) error) error {
	deliveries, err := b.ch.Consume(
		NotificationQueue,
		"notification-worker",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("falha ao iniciar consumidor da fila de notificações: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				return nil
			}

			var msg port.NotificationMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				_ = d.Nack(false, false)
				continue
			}

			if err := handler(ctx, &msg); err != nil {
				log.Printf("Falha ao enviar notificação: %v", err)
				_ = d.Nack(false, true) // reenfileira para tentar novamente
			} else {
				_ = d.Ack(false)
			}
		}
	}
}

func (b *RabbitMQBroker) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.isClosed {
		return nil
	}
	b.isClosed = true

	if b.ch != nil {
		_ = b.ch.Close()
	}
	if b.conn != nil {
		_ = b.conn.Close()
	}
	return nil
}

var _ port.QueueProducer = (*RabbitMQBroker)(nil)
var _ port.QueueConsumer = (*RabbitMQBroker)(nil)
