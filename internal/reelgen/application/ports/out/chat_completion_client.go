package out

import (
	"context"
	"koreels/internal/reelgen/domain/entity"
)

type ChatCompletionClient interface {
	Generate(ctx context.Context, prompt string) (*entity.ChatCompletionResponse, error)
}
