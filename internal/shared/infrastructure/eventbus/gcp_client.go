package eventbus

import (
	"context"
	"errors"
	"fmt"

	"koreels/internal/shared/configuration"

	"cloud.google.com/go/pubsub"
	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewGcpClient)

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
