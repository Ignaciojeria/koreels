package eventbus

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"koreels/internal/shared/infrastructure/httpserver"

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

	go func() {
		ctx := context.Background()
		err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			ce := ps.convertPullMessage(subscriptionName, msg)

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
			ps.Start(subscriptionName, processor, receiveSettings)
			return
		}
	}()

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
		if r.Header.Get("X-Goog-Channel-ID") != "" || r.Header.Get("ce-id") != "" {
			ps.handleNativePush(subName, processor, w, r)
			return
		}

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

	carrier := propagation.MapCarrier{}
	for k, v := range ce.Extensions() {
		if strVal, ok := v.(string); ok {
			carrier[k] = strVal
		}
	}
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), carrier)

	w.WriteHeader(processor(ctx, ce))
}
