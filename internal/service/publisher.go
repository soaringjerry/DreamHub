package service

import (
	"context"
)

// Event represents a message to be published.
// It contains a topic for routing and a payload with the data.
type Event struct {
	Topic   string
	Payload map[string]interface{}
}

// EventPublisher defines the interface for publishing events.
// This decouples services from the specific event bus implementation (e.g., Asynq, NATS, gRPC).
type EventPublisher interface {
	// Publish sends an event to the event bus.
	// It returns a unique identifier for the published event/task and an error if it fails.
	Publish(ctx context.Context, event Event) (id string, err error)
}
