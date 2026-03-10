package fuegoapi

import (
	"errors"
	"fmt"
	"strings"

	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
)

var _ = ioc.Register(NewPostGenerateScenes)

const (
	MaxWords      = 60
	MaxCharacters = 400
)

// NewPostGenerateScenes registers the endpoint to generate scenes using Qwen structured outputs
func NewPostGenerateScenes(s *httpserver.Server, uc in.GenerateScenesExecutor) {
	fuegofw.Post(s.Manager, "/reelgen/scenes",
		func(c fuegofw.ContextWithBody[in.GenerateScenesRequest]) (in.GenerateScenesResponse, error) {
			apiKey := c.Header("x-api-key")
			if apiKey == "" {
				return in.GenerateScenesResponse{}, fuegofw.BadRequestError{
					Err:    errors.New("missing x-api-key header"),
					Title:  "Missing API Key",
					Detail: "The x-api-key header is required to process this request.",
				}
			}

			body, err := c.Body()
			if err != nil {
				return in.GenerateScenesResponse{}, err
			}

			charCount := len([]rune(body.ScriptText))
			if charCount > MaxCharacters {
				return in.GenerateScenesResponse{}, fuegofw.BadRequestError{
					Err:    fmt.Errorf("script text exceeds maximum allowed characters: %d", charCount),
					Title:  "Script Too Long",
					Detail: fmt.Sprintf("The script text contains %d characters. The maximum allowed is %d characters.", charCount, MaxCharacters),
				}
			}

			wordCount := len(strings.Fields(body.ScriptText))
			if wordCount > MaxWords {
				return in.GenerateScenesResponse{}, fuegofw.BadRequestError{
					Err:    fmt.Errorf("script text exceeds maximum allowed words: %d", wordCount),
					Title:  "Script Too Long",
					Detail: fmt.Sprintf("The script text contains %d words. The maximum allowed is %d words.", wordCount, MaxWords),
				}
			}

			body.ProviderAPIKey = apiKey
			return uc.Execute(c.Context(), body)
		})
}
