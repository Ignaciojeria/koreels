package chatcompletion

import (
	"fmt"

	"koreels/internal/reelgen/adapter/out/geminiapi"
	"koreels/internal/reelgen/adapter/out/qwenapi"
	"koreels/internal/reelgen/application/ports/out"
	"koreels/internal/shared/configuration"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewChatCompletionClientFactory)

func NewChatCompletionClientFactory(
	conf configuration.Conf,
	qwen *qwenapi.ChatCompletionClient,
	gemini *geminiapi.ChatCompletionClient,
) (out.ChatCompletionClient, error) {
	switch conf.SCENE_GENERATOR_PROVIDER {
	case "gemini":
		return gemini, nil
	case "qwen":
		return qwen, nil
	default:
		// Default to qwen if not specified
		if conf.SCENE_GENERATOR_PROVIDER == "" {
			return qwen, nil
		}
		return nil, fmt.Errorf("unknown SCENE_GENERATOR_PROVIDER: %s", conf.SCENE_GENERATOR_PROVIDER)
	}
}
