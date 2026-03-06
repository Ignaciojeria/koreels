package fuegoapi

import (
	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
)

var _ = ioc.Register(NewPostGenerateScenes)

// NewPostGenerateScenes registers the endpoint to generate scenes using Qwen structured outputs
func NewPostGenerateScenes(s *httpserver.Server, uc in.GenerateScenesExecutor) {
	fuegofw.Post(s.Manager, "/reelgen/scenes",
		func(c fuegofw.ContextWithBody[in.GenerateScenesRequest]) (in.GenerateScenesResponse, error) {
			body, err := c.Body()
			if err != nil {
				return in.GenerateScenesResponse{}, err
			}
			return uc.Execute(c.Context(), body)
		})
}
