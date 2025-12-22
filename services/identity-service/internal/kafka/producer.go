package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Event struct {
	EventID   string                 `json:"event_id"`
	EventType string                 `json:"event_type"`
	UserID    uuid.UUID              `json:"user_id"`
	Email     string                 `json:"email"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}

	return &Producer{writer: writer}
}

// PublishEvent publishes an event to Kafka
func (p *Producer) PublishEvent(ctx context.Context, event *Event) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	message := kafka.Message{
		Key:   []byte(event.UserID.String()),
		Value: eventJSON,
		Time:  time.Now(),
	}

	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		log.Printf("Failed to publish event: %v", err)
		return err
	}

	log.Printf("Published event: %s for user: %s", event.EventType, event.UserID)
	return nil
}

// Close closes the Kafka writer
func (p *Producer) Close() error {
	return p.writer.Close()
}
