package storage

import (
	"context"

	"koreels/internal/shared/configuration"

	"cloud.google.com/go/storage"
	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewGCSClient)

// NewGCSClient creates a GCS client for uploads (e.g. reelgen audio). Uses default credentials (GOOGLE_APPLICATION_CREDENTIALS or workload identity).
func NewGCSClient(conf configuration.Conf) (*storage.Client, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	_ = conf // reserved for bucket name etc.
	return client, nil
}
