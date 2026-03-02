# consumer

> Inbound EventBus consumer pattern with CloudEvents

## app/adapter/in/eventbus/consumer.go

```go
package eventbus

import (
	"context"
	"log/slog"
	"net/http"

	"archetype/app/shared/infrastructure/eventbus"

	"github.com/Ignaciojeria/ioc"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

var _ = ioc.Register(NewTemplateConsumer)

type TemplateMessage struct {
	ID string `json:"id"`
}

type TemplateConsumer struct {
	subscriber eventbus.Subscriber
}

func NewTemplateConsumer(sub eventbus.Subscriber) (*TemplateConsumer, error) {
	c := &TemplateConsumer{
		subscriber: sub,
	}

	processor := func(ctx context.Context, event cloudevents.Event) int {
		slog.Info("CloudEvent received", "id", event.ID(), "type", event.Type())

		var payload TemplateMessage
		if err := event.DataAs(&payload); err != nil {
			slog.Error("failed_to_unmarshal_cloudevent", "error", err.Error())
			// Invalid payload -> ACK (do not retry infinite loops)
			return http.StatusAccepted
		}

		slog.Info("Successfully processed payload", "id", payload.ID)
		// Process core business logic here...

		// Returning 200 tells the broker to ACK the message
		return http.StatusOK
	}

	// This starts the listening process in background via PULL and binds the PUSH http route.
	// You might want to parameterize "template_topic_or_hook" with an environment variable.
	go c.subscriber.Start("template_topic_or_hook", processor, eventbus.ReceiveSettings{MaxOutstandingMessages: 3})

	return c, nil
}
```

---

## Unit tests

When creating a new component, generate tests following this pattern:

### app/adapter/in/eventbus/consumer_test.go

```go
package eventbus

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"archetype/app/shared/infrastructure/eventbus"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type mockSubscriber struct {
	processor eventbus.MessageProcessor
	mu        sync.Mutex
}

func (m *mockSubscriber) Start(subscriptionName string, processor eventbus.MessageProcessor, receiveSettings eventbus.ReceiveSettings) error {
	m.mu.Lock()
	m.processor = processor
	m.mu.Unlock()
	return nil
}

func (m *mockSubscriber) getProcessor() eventbus.MessageProcessor {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.processor
}

func TestNewTemplateConsumer(t *testing.T) {
	mockSub := &mockSubscriber{}
	c, err := NewTemplateConsumer(mockSub)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if c == nil {
		t.Fatal("expected consumer, got nil")
	}

	// wait max 1 sec for go routine
	for i := 0; i < 100; i++ {
		if mockSub.getProcessor() != nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if mockSub.getProcessor() == nil {
		t.Fatal("expected processor to be registered")
	}

	// Test processor with invalid payload
	ce := cloudevents.NewEvent()
	ce.SetID("123")
	ce.SetType("test.type")
	ce.SetSource("test.source")

	// Force invalid JSON via byte array matching expected errors.
	ce.SetData(cloudevents.ApplicationJSON, map[string]interface{}{})
	// Injecting unparsable format
	ce.DataEncoded = []byte(`{"invalid":json}`)

	status := mockSub.getProcessor()(context.Background(), ce)
	if status != http.StatusAccepted {
		t.Errorf("expected status %d for invalid json, got %d", http.StatusAccepted, status)
	}

	// Test processor with valid payload via struct mapping
	ce.SetData(cloudevents.ApplicationJSON, map[string]string{"id": "test-id"})
	status = mockSub.getProcessor()(context.Background(), ce)
	if status != http.StatusOK {
		t.Errorf("expected status %d for valid json, got %d", http.StatusOK, status)
	}
}
```
