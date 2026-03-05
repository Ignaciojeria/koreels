package eventbus

import (
	"errors"

	"koreels/internal/shared/configuration"

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
