# publisher

> Outbound EventBus publisher adapter

## app/adapter/out/eventbus/publisher.go

```go
package eventbus

import (
	"context"
	"fmt"

	"archetype/app/application/ports/out"
	"archetype/app/shared/infrastructure/eventbus"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewTemplatePublisher)

type templatePublisher struct {
	publisher eventbus.Publisher
}

// NewTemplatePublisher returns an implementation of ports/out.DomainEventPublisher.
func NewTemplatePublisher(publisher eventbus.Publisher) (out.DomainEventPublisher, error) {
	if publisher == nil {
		return nil, fmt.Errorf("publisher dependency is nil")
	}
	return &templatePublisher{
		publisher: publisher,
	}, nil
}

func (p *templatePublisher) Publish(ctx context.Context, e out.Event) error {
	domainEvent, ok := e.(eventbus.DomainEvent)
	if !ok {
		return fmt.Errorf("event must implement eventbus.DomainEvent for CloudEvents serialization")
	}
	request := eventbus.PublishRequest{
		Topic: "your-topic-name",
		Event: domainEvent,
	}
	return p.publisher.Publish(ctx, request)
}
```

---

## Unit tests

When creating a new component, generate tests following this pattern:

### app/adapter/out/eventbus/publisher_test.go

```go
package eventbus

import (
	"context"
	"errors"
	"testing"

	"archetype/app/shared/infrastructure/eventbus"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/stretchr/testify/assert"
)

// MockPublisher wraps around the eventbus.Publisher interface for testing
type MockPublisher struct {
	PublishFunc func(ctx context.Context, request eventbus.PublishRequest) error
}

func (m *MockPublisher) Publish(ctx context.Context, request eventbus.PublishRequest) error {
	return m.PublishFunc(ctx, request)
}

// MockDomainEvent simulates a domain event (implements both ports/out.Event and eventbus.DomainEvent)
type MockDomainEvent struct {
	ID string
}

func (m MockDomainEvent) EventType() string {
	return "mock.event"
}

func (m MockDomainEvent) ToCloudEvent() cloudevents.Event {
	e := cloudevents.NewEvent()
	e.SetID(m.ID)
	return e
}

func TestNewTemplatePublisher(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockPub := &MockPublisher{}
		pub, err := NewTemplatePublisher(mockPub)

		assert.NoError(t, err)
		assert.NotNil(t, pub)
	})

	t.Run("nil dependency", func(t *testing.T) {
		pub, err := NewTemplatePublisher(nil)

		assert.Error(t, err)
		assert.Nil(t, pub)
		assert.Equal(t, "publisher dependency is nil", err.Error())
	})
}

func TestTemplatePublisher_Publish(t *testing.T) {
	var capturedRequest eventbus.PublishRequest
	mockPub := &MockPublisher{
		PublishFunc: func(ctx context.Context, request eventbus.PublishRequest) error {
			capturedRequest = request
			if request.Topic == "error-topic" {
				return errors.New("publish failed")
			}
			return nil
		},
	}

	pub, _ := NewTemplatePublisher(mockPub)
	ctx := context.Background()

	t.Run("successful publish", func(t *testing.T) {
		event := MockDomainEvent{ID: "test-id-123"}

		err := pub.Publish(ctx, event)

		assert.NoError(t, err)
		assert.Equal(t, "your-topic-name", capturedRequest.Topic)
		assert.Equal(t, event, capturedRequest.Event)
	})

	t.Run("publish error passthrough", func(t *testing.T) {
		// Mock error using a specific topic trigger since the actual implementation hardcodes topic.
		// A real test would likely inject the topic depending on the implementation. Let's just mock
		// it globally for the abstract implementation:
		mockPubErr := &MockPublisher{
			PublishFunc: func(ctx context.Context, request eventbus.PublishRequest) error {
				return errors.New("upstream failure")
			},
		}

		pubErr, _ := NewTemplatePublisher(mockPubErr)
		event := MockDomainEvent{ID: "test-id-124"}

		err := pubErr.Publish(ctx, event)
		assert.ErrorContains(t, err, "upstream failure")
	})
}
```
