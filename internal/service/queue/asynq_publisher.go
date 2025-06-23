package queue

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/soaringjerry/dreamhub/internal/service"
)

// AsynqPublisher is an implementation of the EventPublisher interface using Asynq.
type AsynqPublisher struct {
	client *asynq.Client
}

// NewAsynqPublisher creates a new AsynqPublisher.
func NewAsynqPublisher(client *asynq.Client) service.EventPublisher {
	return &AsynqPublisher{
		client: client,
	}
}

// Publish sends an event to the Asynq queue.
// It creates a new task with a type derived from the event topic.
func (p *AsynqPublisher) Publish(ctx context.Context, event service.Event) (string, error) {
	// We can use the event topic to define the Asynq task type.
	// For example, "document.embedding.request" becomes the task type.
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return "", err // Failed to serialize payload
	}
	task := asynq.NewTask(event.Topic, payloadBytes)

	// Enqueue the task.
	taskInfo, err := p.client.EnqueueContext(ctx, task)
	if err != nil {
		return "", err
	}

	// Return the task ID.
	return taskInfo.ID, nil
}
