# eventbus-strategy

> EventBus strategy pattern - interfaces, middleware, and types

## app/shared/infrastructure/eventbus/strategy.go

```go
package eventbus

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// MessageProcessor is the unified callback signature used by ALL brokers.
//
// Return contract:
//   - status < 500  -> ACK  (message handled, do not retry)
//   - status >= 500 -> NACK (retry)
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
```

---

## app/shared/infrastructure/eventbus/factory.go

```go
package eventbus

import (
	"errors"

	"archetype/app/shared/configuration"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewPublisherFactory)
var _ = ioc.Register(NewSubscriberFactory)

// NewPublisherFactory inspects the active EVENT_BROKER and returns the exact valid instance.
func NewPublisherFactory(conf configuration.Conf, natsPub *NatsPublisher, gcpPub *GcpPublisher) (Publisher, error) {
	if conf.EVENT_BROKER == "gcp" {
		if gcpPub == nil {
			return nil, errors.New("gcp broker selected but GcpPublisher failed to initialize")
		}
		return gcpPub, nil
	}
	// Fallback to local NATS default
	if natsPub == nil {
		return nil, errors.New("nats broker selected but NatsPublisher failed to initialize")
	}
	return natsPub, nil
}

// NewSubscriberFactory inspects the active EVENT_BROKER and returns the exact valid instance.
func NewSubscriberFactory(conf configuration.Conf, natsSub *NatsSubscriber, gcpSub *GcpSubscriber) (Subscriber, error) {
	if conf.EVENT_BROKER == "gcp" {
		if gcpSub == nil {
			return nil, errors.New("gcp broker selected but GcpSubscriber failed to initialize")
		}
		return gcpSub, nil
	}
	// Fallback to local NATS default
	if natsSub == nil {
		return nil, errors.New("nats broker selected but NatsSubscriber failed to initialize")
	}
	return natsSub, nil
}
```
