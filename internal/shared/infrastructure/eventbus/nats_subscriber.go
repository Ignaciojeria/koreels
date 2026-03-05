package eventbus

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"koreels/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var _ = ioc.Register(NewNatsSubscriber)

type NatsSubscriber struct {
	client        *NatsClient
	server        *httpserver.Server
	subscriptions []*nats.Subscription
}

func NewNatsSubscriber(client *NatsClient, srv *httpserver.Server) (*NatsSubscriber, error) {
	if client == nil {
		return nil, nil // Disabled via configuration
	}
	return &NatsSubscriber{
		client: client,
		server: srv,
	}, nil
}

func (s *NatsSubscriber) Start(subscriptionName string, processor MessageProcessor, receiveSettings ReceiveSettings) error {
	sub, err := s.client.Connection.Subscribe(subscriptionName, func(m *nats.Msg) {
		s.processMessageAsCloudEvent(m, processor)
	})

	if err != nil {
		return err
	}

	s.subscriptions = append(s.subscriptions, sub)
	return nil
}

func (s *NatsSubscriber) processMessageAsCloudEvent(m *nats.Msg, processor MessageProcessor) {
	ctx := context.Background()

	var ce cloudevents.Event
	err := json.Unmarshal(m.Data, &ce)

	if err != nil {
		log.Printf("invalid cloudevent on NATS: %v", err)
		return
	}

	carrier := propagation.MapCarrier{}
	for k, v := range ce.Extensions() {
		if strVal, ok := v.(string); ok {
			carrier[k] = strVal
		}
	}
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	statusCode := processor(ctx, ce)
	if statusCode != http.StatusOK {
		log.Printf("Processor failed NATS message with status %d", statusCode)
	}
}
