package geminichat

import (
	"koreels/internal/reelgen/adapter/out/geminiapi"
	"koreels/internal/reelgen/application/ports/out"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewProvider)

func NewProvider(gemini *geminiapi.ChatCompletionClient) out.ChatCompletionClient {
	return gemini
}
