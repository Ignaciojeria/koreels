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

func NewGcpPublisher(c *pubsub.Client) (*GcpPublisher, error) {
	if c == nil {
		return nil, nil
	}
	return &GcpPublisher{client: c}, nil
}

func (p *GcpPublisher) Publish(ctx context.Context, request PublishRequest) error {
	ce := request.Event.ToCloudEvent()

	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	for k, v := range carrier {
		ce.SetExtension(k, v)
	}

	bytes, err := json.Marshal(ce)
	if err != nil {
		return fmt.Errorf("cloudevent marshal error: %w", err)
	}

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
