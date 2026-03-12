package fuegoapi

import (
	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
)

var _ = ioc.Register(NewPostGenerateAudio)

// NewPostGenerateAudio registers the endpoint to generate audio for beats
func NewPostGenerateAudio(s *httpserver.Server, uc in.GenerateAudioExecutor) {
	fuegofw.Post(s.Manager, "/reelgen/audio",
		func(c fuegofw.ContextWithBody[in.GenerateAudioRequest]) (in.GenerateAudioResponse, error) {
			// Optional api key parsing if needed:
			apiKey := c.Header("x-api-key")

			body, err := c.Body()
			if err != nil {
				return in.GenerateAudioResponse{}, err
			}
			body.ProviderAPIKey = apiKey

			return uc.Execute(c.Context(), body)
		})
}
