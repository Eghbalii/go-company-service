package port

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventType classifies a domain event.
type EventType string

const (
	EventCompanyCreated EventType = "company.created"
	EventCompanyUpdated EventType = "company.updated"
	EventCompanyDeleted EventType = "company.deleted"
)

// Event is the canonical envelope published to the event broker.
type Event struct {
	EventID   uuid.UUID       `json:"event_id"`
	EventType EventType       `json:"event_type"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// EventPublisher is the contract for async event emission.
type EventPublisher interface {
	Publish(ctx context.Context, event *Event) error
	Close() error
}
