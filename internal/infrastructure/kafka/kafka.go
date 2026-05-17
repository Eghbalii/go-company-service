package kafka

import (
	"github.com/eghbalii/go-company-service/pkg/config"
	"github.com/segmentio/kafka-go"
)

// NewWriter creates a kafka-go Writer for the configured topic and brokers.
// Async=true means Write returns as soon as the message is buffered (fire-and-forget);
// Async=false waits for broker acknowledgement.
func NewWriter(cfg config.Kafka) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.Topic,
		Async:    cfg.Async,
		Balancer: &kafka.LeastBytes{},
		// Retry policy: 3 attempts, avoids dropping events on transient failures.
		MaxAttempts: 3,
	}
}
