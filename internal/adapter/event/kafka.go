package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eghbalii/go-company-service/internal/domain/port"
	"github.com/segmentio/kafka-go"
)

// KafkaPublisher implements port.EventPublisher using a kafka-go Writer.
type KafkaPublisher struct {
	writer *kafka.Writer
}

// NewKafkaPublisher creates a KafkaPublisher wrapping the given Writer.
func NewKafkaPublisher(writer *kafka.Writer) *KafkaPublisher {
	return &KafkaPublisher{writer: writer}
}

// Publish serialises the event and sends it to Kafka.
// When the underlying Writer is async, this returns immediately after buffering.
func (p *KafkaPublisher) Publish(ctx context.Context, event *port.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.EventID.String()),
		Value: payload,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("write kafka message: %w", err)
	}

	return nil
}

// Close shuts the underlying Kafka writer and flushes any buffered messages.
func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}
