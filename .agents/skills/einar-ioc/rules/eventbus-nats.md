# eventbus-nats

> NATS client, publisher, and subscriber implementation

## app/shared/infrastructure/eventbus/nats_client.go

```go
package eventbus

import (
	"log"
	"time"

	"archetype/app/shared/configuration"

	"github.com/Ignaciojeria/ioc"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

var _ = ioc.Register(NewNatsClient)

// NatsClient encapsulates an in-memory NATS server and the connection to it.
// To use a remote server instead, ignore the EmbeddedServer initialization
// and configure nats.Connect with the real URL.
type NatsClient struct {
	EmbeddedServer *server.Server
	Connection     *nats.Conn
}

// NewNatsClient sets up a lightweight in-memory NATS server natively
// within the Go application and connects a client to it. Very useful for local
// development and tests without requiring heavy infrastructure.
func NewNatsClient(conf configuration.Conf) (*NatsClient, error) {
	if conf.EVENT_BROKER != "nats" {
		return nil, nil
	}

	// 1) Spin up embedded NATS on a random local port or predefined one
	opts := &server.Options{
		// -1 dynamically picks a port, but if we need a static one for
		// explicit monitoring we can define it. -1 is safest for local testing.
		Port: -1,
	}

	embedded, err := server.NewServer(opts)
	if err != nil {
		return nil, err
	}

	// 2) Start in background
	go embedded.Start()

	if !embedded.ReadyForConnections(5 * time.Second) {
		return nil, err
	}

	log.Printf("Embedded NATS Server running on %s", embedded.ClientURL())

	// 3) Connect a client locally
	conn, err := nats.Connect(embedded.ClientURL())
	if err != nil {
		return nil, err
	}

	return &NatsClient{
		EmbeddedServer: embedded,
		Connection:     conn,
	}, nil
}
```

---

## app/shared/infrastructure/eventbus/nats_publisher.go

```go
package eventbus

import (
	"context"
	"encoding/json"

	"github.com/Ignaciojeria/ioc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var _ = ioc.Register(NewNatsPublisher)

// NatsPublisher implements Publisher matching GCP semantics but over NATS core
type NatsPublisher struct {
	client *NatsClient
}

// NewNatsPublisher creates a Publisher using the Embedded NATS client connection.
func NewNatsPublisher(client *NatsClient) *NatsPublisher {
	if client == nil {
		return nil
	}
	return &NatsPublisher{
		client: client,
	}
}

// Publish takes a DomainEvent, converts it to a standard CloudEvent, serializes it,
// grabs the OTel active span, injects it into CloudEvent Extensions, and
// finally publishes it over the embedded NATS broker.
func (p *NatsPublisher) Publish(ctx context.Context, request PublishRequest) error {

	// 1) Translate domain to CloudEvent wrapper
	ce := request.Event.ToCloudEvent()
	if ce.ID() == "" {
		ce.SetID("nats-" + request.Topic)
	}
	ce.SetSource("ioc-service")

	// 2) Inject OpenTelemetry context for distributed tracing continuity
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	for k, v := range carrier {
		ce.SetExtension(k, v)
	}

	// 3) Serialize CloudEvent object strictly to JSON bytes
	body, err := json.Marshal(ce)
	if err != nil {
		return err
	}

	// 4) Publish down into NATS subject directly
	return p.client.Connection.Publish(request.Topic, body)
}
```

---

## app/shared/infrastructure/eventbus/nats_subscriber.go

```go
package eventbus

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"archetype/app/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var _ = ioc.Register(NewNatsSubscriber)

// NatsSubscriber implements the Subscriber contract fetching from NATS locally.
type NatsSubscriber struct {
	client        *NatsClient
	server        *httpserver.Server
	subscriptions []*nats.Subscription
}

// NewNatsSubscriber initializes the struct using the in-memory NATS client
func NewNatsSubscriber(client *NatsClient, srv *httpserver.Server) (*NatsSubscriber, error) {
	if client == nil {
		return nil, nil // Disabled via configuration
	}
	return &NatsSubscriber{
		client: client,
		server: srv,
	}, nil
}

// Start adds a subscription natively inside the NATS broker
func (s *NatsSubscriber) Start(subscriptionName string, processor MessageProcessor, receiveSettings ReceiveSettings) error {
	// Create a NATS local QueueSubscription for load-balancing semantics if needed, or normal sub.
	sub, err := s.client.Connection.Subscribe(subscriptionName, func(m *nats.Msg) {
		s.processMessageAsCloudEvent(m, processor)
	})

	if err != nil {
		return err
	}

	s.subscriptions = append(s.subscriptions, sub)
	return nil
}

// processMessageAsCloudEvent unwraps the NATS payload exactly like GCP PULL wrappers do.
func (s *NatsSubscriber) processMessageAsCloudEvent(m *nats.Msg, processor MessageProcessor) {
	ctx := context.Background()

	var ce cloudevents.Event
	err := json.Unmarshal(m.Data, &ce)

	if err != nil {
		// Log errors similar to how we handled Google Cloud Events
		log.Printf("invalid cloudevent on NATS: %v", err)
		return
	}

	// Unpack OTel propagating context
	carrier := propagation.MapCarrier{}
	for k, v := range ce.Extensions() {
		if strVal, ok := v.(string); ok {
			carrier[k] = strVal
		}
	}
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	// In memory, we act like handlers responding 200 strictly.
	statusCode := processor(ctx, ce)
	if statusCode != http.StatusOK {
		log.Printf("Processor failed NATS message with status %d", statusCode)
	}
}
```

---

## Unit tests

When creating a new component, generate tests following this pattern:

### app/shared/infrastructure/eventbus/nats_client_test.go

```go
package eventbus

import (
	"context"
	"net/http"
	"testing"
	"time"

	"archetype/app/shared/configuration"
	"archetype/app/shared/infrastructure/httpserver"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/stretchr/testify/assert"
)

// NatsDummyEvent for test integration
type NatsDummyEvent struct {
	Message string `json:"message"`
}

func (d NatsDummyEvent) ToCloudEvent() cloudevents.Event {
	e := cloudevents.NewEvent()
	e.SetID("test-id")
	e.SetType("test.dummy.event")
	e.SetData(cloudevents.ApplicationJSON, d)
	return e
}

func TestNatsIntegrationSuite(t *testing.T) {
	conf := configuration.Conf{
		PORT:         "0",
		PROJECT_NAME: "test",
		VERSION:      "1.0",
		EVENT_BROKER: "nats",
	}

	// 1) Initialize the Embedded Cluster
	client, err := NewNatsClient(conf)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	defer client.EmbeddedServer.Shutdown()
	defer client.Connection.Close()

	// 2) Prepare the tools
	pub := NewNatsPublisher(client)
	srv := &httpserver.Server{}
	sub, _ := NewNatsSubscriber(client, srv)

	// 3) Hook up a handler capturing success
	received := make(chan bool, 1)

	sub.Start("test-topic", func(ctx context.Context, e cloudevents.Event) int {
		var payload NatsDummyEvent
		if err := e.DataAs(&payload); err == nil {
			if payload.Message == "hello from memory" {
				received <- true
				return http.StatusOK
			}
		}
		return http.StatusBadRequest
	}, ReceiveSettings{})

	// 4) Publish to Memory
	event := NatsDummyEvent{Message: "hello from memory"}
	req := PublishRequest{
		Topic: "test-topic",
		Event: event,
	}

	err = pub.Publish(context.Background(), req)
	assert.NoError(t, err)

	// 5) Verify asynchronous delivery natively inside Go
	select {
	case <-received:
		// Success!
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for NATS message")
	}
}
```
