package out

import (
	"context"
	"koreels/internal/reelgen/domain/entity"
)

type ChatCompletionClient interface {
	Generate(ctx context.Context, systemPrompt, userPrompt string, responseFormat interface{}, apiKey string) (*entity.ChatCompletionResponse, error)
}
