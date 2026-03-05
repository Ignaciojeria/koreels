package eventbus

import (
	"context"
	"encoding/json"

	"github.com/Ignaciojeria/ioc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var _ = ioc.Register(NewNatsPublisher)

type NatsPublisher struct {
	client *NatsClient
}

func NewNatsPublisher(client *NatsClient) *NatsPublisher {
	if client == nil {
		return nil
	}
	return &NatsPublisher{
		client: client,
	}
}

func (p *NatsPublisher) Publish(ctx context.Context, request PublishRequest) error {
	ce := request.Event.ToCloudEvent()
	if ce.ID() == "" {
		ce.SetID("nats-" + request.Topic)
	}
	ce.SetSource("ioc-service")

	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	for k, v := range carrier {
		ce.SetExtension(k, v)
	}

	body, err := json.Marshal(ce)
	if err != nil {
		return err
	}

	return p.client.Connection.Publish(request.Topic, body)
}
