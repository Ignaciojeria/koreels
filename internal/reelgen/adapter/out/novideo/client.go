package novideo

import (
	"context"
	"fmt"

	"koreels/internal/reelgen/application/ports/out"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewNoOpVideoClient)

type noOpVideoClient struct{}

func NewNoOpVideoClient() out.VideoGenerationClient {
	return &noOpVideoClient{}
}

func (c *noOpVideoClient) GenerateVideo(_ context.Context, _ string, _ string, _ int, _ string) (*out.VideoGenerationResult, error) {
	return nil, fmt.Errorf("video generation not available in CLI mode")
}
