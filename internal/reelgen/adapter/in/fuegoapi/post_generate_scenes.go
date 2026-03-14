package fuegoapi

import (
	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
)

var _ = ioc.Register(NewPostGenerateScenes)

// NewPostGenerateScenes registra POST /reelgen/scenes: agrupa beats en escenas ≥ 4s y devuelve el plan de escenas.
func NewPostGenerateScenes(s *httpserver.Server, uc in.GenerateScenesExecutor) {
	fuegofw.Post(s.Manager, "/reelgen/scenes",
		func(c fuegofw.ContextWithBody[in.GenerateScenesRequest]) (in.GenerateScenesResponse, error) {
			body, err := c.Body()
			if err != nil {
				return in.GenerateScenesResponse{}, err
			}
			body.ProviderAPIKey = c.Header("x-api-key")
			return uc.Execute(c.Context(), body)
		})
}
