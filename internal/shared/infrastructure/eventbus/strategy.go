package eventbus

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// MessageProcessor is the unified callback signature used by ALL brokers.
type MessageProcessor func(ctx context.Context, event cloudevents.Event) int

// ProcessorMiddleware allows you to wrap processors with additional behavior.
type ProcessorMiddleware func(MessageProcessor) MessageProcessor

// ApplyMiddlewares builds a final MessageProcessor with middleware.
func ApplyMiddlewares(
	processor MessageProcessor,
	middlewares ...ProcessorMiddleware,
) MessageProcessor {
	for i := len(middlewares) - 1; i >= 0; i-- {
		processor = middlewares[i](processor)
	}
	return processor
}

// ReceiveSettings configures generic pull settings.
type ReceiveSettings struct {
	MaxOutstandingMessages int
}

// Subscriber is the cross-broker abstraction for reading queue events.
type Subscriber interface {
	Start(subscriptionName string, processor MessageProcessor, receiveSettings ReceiveSettings) error
}

// DomainEvent is anything that can be converted to CloudEvent to be published globally.
type DomainEvent interface {
	ToCloudEvent() cloudevents.Event
}

type PublishRequest struct {
	Topic       string
	OrderingKey string // Optional: for ordering in FIFO queues
	Event       DomainEvent
}

// Publisher is the vendor-neutral interface for sending events.
type Publisher interface {
	Publish(ctx context.Context, request PublishRequest) error
}
