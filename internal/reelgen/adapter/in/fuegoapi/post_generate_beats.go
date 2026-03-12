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

var _ = ioc.Register(NewPostGenerateBeats)

const (
	MaxWords      = 60
	MaxCharacters = 400
)

// NewPostGenerateBeats registers the endpoint to generate beats using Qwen structured outputs
func NewPostGenerateBeats(s *httpserver.Server, uc in.GenerateBeatsExecutor) {
	fuegofw.Post(s.Manager, "/reelgen/beats",
		func(c fuegofw.ContextWithBody[in.GenerateBeatsRequest]) (in.GenerateBeatsResponse, error) {
			apiKey := c.Header("x-api-key")
			if apiKey == "" {
				return in.GenerateBeatsResponse{}, fuegofw.BadRequestError{
					Err:    errors.New("missing x-api-key header"),
					Title:  "Missing API Key",
					Detail: "The x-api-key header is required to process this request.",
				}
			}

			body, err := c.Body()
			if err != nil {
				return in.GenerateBeatsResponse{}, err
			}

			charCount := len([]rune(body.ScriptText))
			if charCount > MaxCharacters {
				return in.GenerateBeatsResponse{}, fuegofw.BadRequestError{
					Err:    fmt.Errorf("script text exceeds maximum allowed characters: %d", charCount),
					Title:  "Script Too Long",
					Detail: fmt.Sprintf("The script text contains %d characters. The maximum allowed is %d characters.", charCount, MaxCharacters),
				}
			}

			wordCount := len(strings.Fields(body.ScriptText))
			if wordCount > MaxWords {
				return in.GenerateBeatsResponse{}, fuegofw.BadRequestError{
					Err:    fmt.Errorf("script text exceeds maximum allowed words: %d", wordCount),
					Title:  "Script Too Long",
					Detail: fmt.Sprintf("The script text contains %d words. The maximum allowed is %d words.", wordCount, MaxWords),
				}
			}

			body.ProviderAPIKey = apiKey
			return uc.Execute(c.Context(), body)
		})
}
