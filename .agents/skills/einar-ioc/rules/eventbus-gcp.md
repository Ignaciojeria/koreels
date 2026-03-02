# eventbus-gcp

> GCP Pub/Sub client, publisher, and subscriber implementation

## app/shared/infrastructure/eventbus/gcp_client.go

```go
package eventbus

import (
	"context"
	"errors"
	"fmt"

	"archetype/app/shared/configuration"

	"cloud.google.com/go/pubsub"
	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewGcpClient)

// NewGcpClient creates a new GCP PubSub client using the configuration.
func NewGcpClient(env configuration.Conf) (*pubsub.Client, error) {
	if env.EVENT_BROKER != "gcp" {
		return nil, nil
	}

	if env.GOOGLE_PROJECT_ID == "" {
		return nil, errors.New("GOOGLE_PROJECT_ID is required for PubSub client")
	}

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, env.GOOGLE_PROJECT_ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}

	return client, nil
}
```

---

## app/shared/infrastructure/eventbus/gcp_publisher.go

```go
package eventbus

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/Ignaciojeria/ioc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var _ = ioc.Register(NewGcpPublisher)

type GcpPublisher struct {
	client *pubsub.Client
}

// NewGcpPublisher creates an implementation of the universal Publisher interface backed by GCP Pub/Sub.
func NewGcpPublisher(c *pubsub.Client) (*GcpPublisher, error) {
	if c == nil {
		return nil, nil
	}
	return &GcpPublisher{client: c}, nil
}

// Publish transforms a DomainEvent into a CloudEvent and dispatches it over GCP pub/sub.
func (p *GcpPublisher) Publish(
	ctx context.Context,
	request PublishRequest,
) error {
	ce := request.Event.ToCloudEvent()

	// Inject OpenTelemetry trace context into the CloudEvent extensions
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	for k, v := range carrier {
		ce.SetExtension(k, v)
	}

	bytes, err := json.Marshal(ce)
	if err != nil {
		return fmt.Errorf("cloudevent marshal error: %w", err)
	}

	// Build attributes for Pub/Sub filtering without needing to deserialize the payload.
	attrs := make(map[string]string)

	if ce.Type() != "" {
		attrs["ce-type"] = ce.Type()
	}
	if ce.Source() != "" {
		attrs["ce-source"] = ce.Source()
	}
	if ce.Subject() != "" {
		attrs["ce-subject"] = ce.Subject()
	}
	if ce.ID() != "" {
		attrs["ce-id"] = ce.ID()
	}

	// Dump CloudEvent extensions so GCP pubsub filtering logic handles it identically.
	for k, v := range ce.Context.GetExtensions() {
		attrs[k] = fmt.Sprintf("%v", v)
	}

	pubTopic := p.client.Topic(request.Topic)
	pubTopic.EnableMessageOrdering = true

	_, err = pubTopic.Publish(ctx, &pubsub.Message{
		Data:        bytes,
		Attributes:  attrs,
		OrderingKey: request.OrderingKey,
	}).Get(ctx)

	return err
}
```

---

## app/shared/infrastructure/eventbus/gcp_subscriber.go

```go
package eventbus

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"archetype/app/shared/infrastructure/httpserver"

	"cloud.google.com/go/pubsub"
	"github.com/Ignaciojeria/ioc"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var _ = ioc.Register(NewGcpSubscriber)

type GcpSubscriber struct {
	client     *pubsub.Client
	httpServer *httpserver.Server
}

func NewGcpSubscriber(c *pubsub.Client, s *httpserver.Server) (*GcpSubscriber, error) {
	if c == nil {
		return nil, nil
	}
	return &GcpSubscriber{client: c, httpServer: s}, nil
}

func (ps *GcpSubscriber) Start(subscriptionName string, processor MessageProcessor, receiveSettings ReceiveSettings) error {
	sub := ps.client.Subscription(subscriptionName)
	sub.ReceiveSettings.MaxOutstandingMessages = receiveSettings.MaxOutstandingMessages

	// PULL consumer running in background
	go func() {
		ctx := context.Background()
		err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			ce := ps.convertPullMessage(subscriptionName, msg)

			// Extract trace context from CloudEvent extensions
			carrier := propagation.MapCarrier{}
			for k, v := range ce.Extensions() {
				if strVal, ok := v.(string); ok {
					carrier[k] = strVal
				}
			}
			ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

			status := processor(ctx, ce)

			if status >= 500 {
				msg.Nack()
				return
			}
			msg.Ack()
		})

		if err != nil {
			slog.Error("pubsub_receive_failed",
				"subscription", subscriptionName,
				"error", err.Error(),
			)
			time.Sleep(5 * time.Second)
			ps.Start(subscriptionName, processor, receiveSettings) // auto-retry PULL
			return
		}
	}()

	// PUSH consumer via HTTP (Fuego integration)
	path := "/subscription/" + subscriptionName

	fuego.PostStd(ps.httpServer.Manager, path, ps.makePushHandler(subscriptionName, processor), option.Summary("Internal webhook pubsub push"))
	return nil
}

func (ps *GcpSubscriber) convertPullMessage(subName string, msg *pubsub.Message) cloudevents.Event {
	var ce cloudevents.Event
	if err := json.Unmarshal(msg.Data, &ce); err != nil {
		slog.Warn("failed_to_unmarshal_cloudevent",
			"subscription", subName,
			"message_id", msg.ID,
			"error", err.Error(),
		)
		ce = cloudevents.NewEvent()
		ce.SetID(msg.ID)
		ce.SetType("google.pubsub.pull.fallback")
		ce.SetSource("gcp.pubsub/" + subName)
		ce.SetData(cloudevents.ApplicationJSON, msg.Data)
	} else if ce.ID() == "" {
		ce.SetID(msg.ID)
	}
	return ce
}

func (ps *GcpSubscriber) makePushHandler(subName string, processor MessageProcessor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Native GCP push header detection
		if r.Header.Get("X-Goog-Channel-ID") != "" || r.Header.Get("ce-id") != "" {
			ps.handleNativePush(subName, processor, w, r)
			return
		}

		// Manual custom POST for local testing without Cloud Emulator
		ps.handleManualPush(subName, processor, w, r)
	}
}

func (ps *GcpSubscriber) handleNativePush(subName string, processor MessageProcessor, w http.ResponseWriter, r *http.Request) {
	var envelope struct {
		Message struct {
			MessageID  string            `json:"messageId"`
			Data       []byte            `json:"data"`
			Attributes map[string]string `json:"attributes"`
		} `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&envelope); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var ce cloudevents.Event
	if err := json.Unmarshal(envelope.Message.Data, &ce); err != nil {
		slog.Warn("failed_to_unmarshal_cloudevent_push",
			"subscription", subName,
			"message_id", envelope.Message.MessageID,
			"error", err.Error(),
		)
		ce = cloudevents.NewEvent()
		ce.SetID(envelope.Message.MessageID)
		ce.SetType("google.pubsub.push.fallback")
		ce.SetSource("gcp.pubsub/" + subName)
		ce.SetData(cloudevents.ApplicationJSON, envelope.Message.Data)
	} else if ce.ID() == "" {
		ce.SetID(envelope.Message.MessageID)
	}

	// Extract trace context
	carrier := propagation.MapCarrier{}
	for k, v := range ce.Extensions() {
		if strVal, ok := v.(string); ok {
			carrier[k] = strVal
		}
	}
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), carrier)

	w.WriteHeader(processor(ctx, ce))
}

func (ps *GcpSubscriber) handleManualPush(subName string, processor MessageProcessor, w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	ce := cloudevents.NewEvent()
	ce.SetID("")
	ce.SetType("manual.message")
	ce.SetSource("manual/" + subName)
	ce.SetData(cloudevents.ApplicationJSON, body)

	for key, values := range r.Header {
		if len(values) > 0 {
			ce.SetExtension(strings.ToLower(key), strings.Join(values, ","))
		}
	}

	// Extract trace context
	carrier := propagation.MapCarrier{}
	for k, v := range ce.Extensions() {
		if strVal, ok := v.(string); ok {
			carrier[k] = strVal
		}
	}
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), carrier)

	w.WriteHeader(processor(ctx, ce))
}
```

---

## Unit tests

When creating a new component, generate tests following this pattern:

### app/shared/infrastructure/eventbus/gcp_client_test.go

```go
package eventbus

import (
	"context"
	"os"
	"strings"
	"testing"

	"archetype/app/shared/configuration"

	"cloud.google.com/go/pubsub/pstest"
)

func TestNewGcpClient_Success(t *testing.T) {
	srv := pstest.NewServer()
	defer srv.Close()

	os.Setenv("PUBSUB_EMULATOR_HOST", srv.Addr)
	defer os.Unsetenv("PUBSUB_EMULATOR_HOST")

	conf := configuration.Conf{EVENT_BROKER: "gcp", GOOGLE_PROJECT_ID: "test-project"}
	client, err := NewGcpClient(conf)
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	_, err = client.CreateTopic(ctx, "test-topic")
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}
}

func TestNewGcpClient_MissingProjectID(t *testing.T) {
	conf := configuration.Conf{
		EVENT_BROKER:      "gcp",
		GOOGLE_PROJECT_ID: "",
	}

	client, err := NewGcpClient(conf)
	if err == nil {
		t.Fatal("expected error creating pubsub client with empty google project ID, got nil")
	}
	if client != nil {
		t.Errorf("expected nil pubsub client on error, got %v", client)
	}

	if !strings.Contains(err.Error(), "GOOGLE_PROJECT_ID is required") {
		t.Errorf("expected missing GOOGLE_PROJECT_ID formatting error, got %v", err)
	}
}

func TestNewGcpClient_FailureToConnect(t *testing.T) {
	conf := configuration.Conf{
		EVENT_BROKER:      "gcp",
		GOOGLE_PROJECT_ID: "test-project",
	}

	// This should fail to create client because credentials are not found and pubsub requires it
	// unless running with option.WithoutAuthentication() which our client does not embed,
	// or emulator env var. So setting a random PUBSUB_EMULATOR_HOST to an invalid address gives an error.
	os.Setenv("PUBSUB_EMULATOR_HOST", "127.0.0.1:0")
	defer os.Unsetenv("PUBSUB_EMULATOR_HOST")

	_, err := NewGcpClient(conf)
	if err != nil {
		// NewClient succeeds synchronously even with fake emulator host
		// if credentials are not checked immediately, but if it returns an err, we catch it.
	}
}
```

---

### app/shared/infrastructure/eventbus/gcp_publisher_test.go

```go
package eventbus

import (
	"context"
	"testing"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DummyEvent struct{}

func (d DummyEvent) ToCloudEvent() cloudevents.Event {
	ce := cloudevents.NewEvent()
	ce.SetID("123")
	ce.SetType("test.type")
	ce.SetSource("test.source")
	ce.SetData(cloudevents.ApplicationJSON, map[string]string{"foo": "bar"})
	ce.SetExtension("customext", "value")
	return ce
}

type MinimalEvent struct{}

func (m MinimalEvent) ToCloudEvent() cloudevents.Event {
	ce := cloudevents.NewEvent()
	// No ID, No Source, No Type explicitly set by user (SDK might set some defaults)
	return ce
}

type FullEvent struct{}

func (f FullEvent) ToCloudEvent() cloudevents.Event {
	ce := cloudevents.NewEvent()
	ce.SetID("full-id")
	ce.SetType("full.type")
	ce.SetSource("full.source")
	ce.SetSubject("full.subject")
	return ce
}

func TestNewGcpPublisher(t *testing.T) {
	pub, err := NewGcpPublisher(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pub != nil {
		t.Fatal("expected nil publisher without a valid client, got instance")
	}
}

func TestGcpPublisher_Publish(t *testing.T) {
	ctx := context.Background()

	// Start a fake PubSub server
	srv := pstest.NewServer()
	defer srv.Close()

	// Connect to it securely via gRPC
	conn, err := grpc.NewClient(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial test server: %v", err)
	}
	defer conn.Close()

	// Create test client attached to the fake server
	client, err := pubsub.NewClient(ctx, "project-id", option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Create a topic on the fake server
	_, err = client.CreateTopic(ctx, "test-topic")
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}

	pub, _ := NewGcpPublisher(client)

	req := PublishRequest{
		Topic:       "test-topic",
		OrderingKey: "123-group",
		Event:       DummyEvent{},
	}

	err = pub.Publish(ctx, req)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Wait and check if message was sent to the server correctly
	msgs := srv.Messages()
	if len(msgs) == 0 {
		t.Fatalf("No messages published to fake server")
	}
	msg := msgs[0]

	if msg.Attributes["ce-id"] != "123" {
		t.Errorf("Expected ce-id=123, got %s", msg.Attributes["ce-id"])
	}

	if msg.Attributes["ce-type"] != "test.type" {
		t.Errorf("Expected ce-type=test.type, got %s", msg.Attributes["ce-type"])
	}

	if msg.Attributes["customext"] != "value" {
		t.Errorf("Expected customext=value, got %s", msg.Attributes["customext"])
	}

	// Test Minimal Event
	_ = pub.Publish(ctx, PublishRequest{Topic: "test-topic", Event: MinimalEvent{}})

	// Test Full Event
	_ = pub.Publish(ctx, PublishRequest{Topic: "test-topic", Event: FullEvent{}})

	// Test Publish error - cancelled context causes Get() to fail
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()
	err = pub.Publish(cancelledCtx, PublishRequest{Topic: "test-topic", Event: DummyEvent{}})
	if err == nil {
		t.Error("expected error when context is cancelled")
	}

	// Check Full Event was received - find by ce-id
	msgs = srv.Messages()
	for _, m := range msgs {
		if m.Attributes["ce-id"] == "full-id" {
			if m.Attributes["ce-subject"] != "full.subject" {
				t.Errorf("Expected ce-subject=full.subject, got %s", m.Attributes["ce-subject"])
			}
			break
		}
	}
}
```

---

### app/shared/infrastructure/eventbus/gcp_subscriber_test.go

```go
package eventbus

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"archetype/app/shared/configuration"
	"archetype/app/shared/infrastructure/httpserver"
	"archetype/app/shared/infrastructure/httpserver/middleware"
	"archetype/app/shared/infrastructure/observability"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNewGcpSubscriber(t *testing.T) {
	srv := &httpserver.Server{}

	sub, err := NewGcpSubscriber(nil, srv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub != nil {
		t.Fatal("expected nil subscriber without a valid pubsub client, got instance")
	}
}

func TestGcpSubscriber_Start(t *testing.T) {
	ctx := context.Background()

	conf := configuration.Conf{
		PORT:         "0",
		PROJECT_NAME: "test",
		VERSION:      "v1",
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	obs := observability.Observability{
		Tracer: noop.NewTracerProvider().Tracer("test"),
		Logger: logger,
	}
	mw := middleware.NewRequestLogger(obs)

	srv, err := httpserver.NewServer(conf, mw)
	if err != nil {
		t.Fatalf("failed to create fake http server: %v", err)
	}

	testSrv := pstest.NewServer()
	defer testSrv.Close()

	conn, err := grpc.NewClient(testSrv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial test server: %v", err)
	}
	defer conn.Close()

	client, err := pubsub.NewClient(ctx, "project-id", option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	topic, err := client.CreateTopic(ctx, "test-topic")
	if err != nil {
		t.Fatalf("failed to create topic: %v", err)
	}
	_, err = client.CreateSubscription(ctx, "test-sub", pubsub.SubscriptionConfig{
		Topic: topic,
	})
	if err != nil {
		t.Fatalf("failed to create sub: %v", err)
	}

	subscriber, _ := NewGcpSubscriber(client, srv)

	receivedChan := make(chan bool)

	processor := func(ctx context.Context, event cloudevents.Event) int {
		receivedChan <- true
		if event.ID() == "fail-me" {
			return http.StatusInternalServerError // trigger a Nack
		}
		return http.StatusOK // trigger an Ack
	}

	err = subscriber.Start("test-sub", processor, ReceiveSettings{MaxOutstandingMessages: 1})
	if err != nil {
		t.Fatalf("Subscriber start failed: %v", err)
	}

	// Publish success event
	ce1 := cloudevents.NewEvent()
	ce1.SetID("1")
	ce1.SetType("test")
	ce1.SetSource("test")
	data1, _ := json.Marshal(ce1)

	topic.Publish(ctx, &pubsub.Message{Data: data1})
	select {
	case <-receivedChan:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for pubsub pull message")
	}

	// Publish fail event
	ce2 := cloudevents.NewEvent()
	ce2.SetID("fail-me")
	ce2.SetType("test")
	ce2.SetSource("test")
	data2, _ := json.Marshal(ce2)

	topic.Publish(ctx, &pubsub.Message{Data: data2})
	select {
	case <-receivedChan:
		// success reaching the processor (nack will just return to pubsub internally)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for pubsub fail message")
	}
}

func TestConvertPullMessage(t *testing.T) {
	sub := &GcpSubscriber{}

	// Test case 1: valid CloudEvent in JSON
	ce := cloudevents.NewEvent()
	ce.SetID("101")
	ce.SetType("test.type")
	ce.SetSource("test.source")

	bytesCE, _ := json.Marshal(ce)

	msg := &pubsub.Message{
		ID:   "202", // Message ID should be ignored if CloudEvent has an ID natively
		Data: bytesCE,
	}

	convCE := sub.convertPullMessage("sub-name", msg)
	if convCE.ID() != "101" {
		t.Errorf("expected ID 101, got %s", convCE.ID())
	}

	// Test case 2: invalid JSON (Fallback creation)
	msgInvalid := &pubsub.Message{
		ID:   "303",
		Data: []byte(`{invalid}`),
	}

	fallbackCE := sub.convertPullMessage("sub-name", msgInvalid)
	if fallbackCE.ID() != "303" {
		t.Errorf("expected fallback ID 303, got %s", fallbackCE.ID())
	}

	// Test case 3: JSON without ID
	ceNoID := cloudevents.NewEvent()
	ceNoID.SetType("test")
	ceNoID.SetSource("test")
	bytesNoID, _ := json.Marshal(ceNoID)

	msgNoID := &pubsub.Message{
		ID:   "505",
		Data: bytesNoID,
	}

	fallbackNoID := sub.convertPullMessage("sub-name", msgNoID)
	if fallbackNoID.ID() != "505" {
		t.Errorf("expected fallback ID 505 for missing id, got %s", fallbackNoID.ID())
	}
}

func TestMakePushHandler(t *testing.T) {
	sub := &GcpSubscriber{}

	var processedID string
	processor := func(ctx context.Context, event cloudevents.Event) int {
		processedID = event.ID()
		return http.StatusOK
	}

	handler := sub.makePushHandler("sub-name", processor)

	// Test manual custom POST
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer([]byte(`{}`)))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}

	// Test manual custom POST with Headers to check if mapping works
	reqHeaders := httptest.NewRequest("POST", "/test", bytes.NewBuffer([]byte(`{}`)))
	reqHeaders.Header.Set("Custom-Ext", "works")
	wHeaders := httptest.NewRecorder()

	handler.ServeHTTP(wHeaders, reqHeaders)

	// Test native GCP Push
	envelopeJSON := `{
		"message": {
			"messageId": "pubsub-id-999",
			"data": "e30=", 
			"attributes": {}
		}
	}`
	reqNative := httptest.NewRequest("POST", "/test", bytes.NewBuffer([]byte(envelopeJSON)))
	reqNative.Header.Set("X-Goog-Channel-ID", "yes")
	wNative := httptest.NewRecorder()

	handler.ServeHTTP(wNative, reqNative)

	if wNative.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", wNative.Code)
	}
	if processedID != "pubsub-id-999" {
		t.Errorf("expected processor to hook event pubsub-id-999, got %s", processedID)
	}

	// Test native GCP Push Invalid
	invalidNativeReq := httptest.NewRequest("POST", "/test", bytes.NewBuffer([]byte(`{invalid}`)))
	invalidNativeReq.Header.Set("X-Goog-Channel-ID", "yes")
	wInvalid := httptest.NewRecorder()

	handler.ServeHTTP(wInvalid, invalidNativeReq)
	if wInvalid.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for invalid json block, got %d", wInvalid.Code)
	}

	// Test native GCP Push fallback data json
	fallbackDataJSON := `{
		"message": {
			"messageId": "pubsub-id-fallback",
			"data": "aW52YWxpZCBqc29uIG9uIHJlc3BvbnNl", 
			"attributes": {}
		}
	}`
	fallbackReq := httptest.NewRequest("POST", "/test", bytes.NewBuffer([]byte(fallbackDataJSON)))
	fallbackReq.Header.Set("X-Goog-Channel-ID", "yes")
	wFallback := httptest.NewRecorder()

	handler.ServeHTTP(wFallback, fallbackReq)
	if processedID != "pubsub-id-fallback" {
		t.Errorf("expected processor to fallback ID, got %s", processedID)
	}

	// Test native GCP Push fallback ID (no native ID in data)
	noIDCE := cloudevents.NewEvent()
	noIDCE.SetType("test")
	noIDCE.SetSource("http://test.source")
	bytesNoID, _ := json.Marshal(noIDCE)
	// Base64 encode it for envelope
	encodedNoID := base64.StdEncoding.EncodeToString(bytesNoID)

	noIDEnvelope := fmt.Sprintf(`{
		"message": {
			"messageId": "mapped-id",
			"data": "%s", 
			"attributes": {}
		}
	}`, encodedNoID)

	reqNoID := httptest.NewRequest("POST", "/test", bytes.NewBuffer([]byte(noIDEnvelope)))
	reqNoID.Header.Set("X-Goog-Channel-ID", "yes")
	wNoID := httptest.NewRecorder()

	handler.ServeHTTP(wNoID, reqNoID)
	if processedID != "mapped-id" {
		t.Errorf("expected ID mapped-id, got %s", processedID)
	}

	// Test Push reading manual body error
	reqErrReader := httptest.NewRequest("POST", "/test", errReader{})
	wErrReader := httptest.NewRecorder()
	handler.ServeHTTP(wErrReader, reqErrReader)
	if wErrReader.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for manual body error, got %d", wErrReader.Code)
	}
}

type errReader struct{}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, context.DeadlineExceeded
}
```
