package fuegoapi

import (
	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
)

var _ = ioc.Register(NewPostGenerateKeyFrames)

func NewPostGenerateKeyFrames(s *httpserver.Server, uc in.GenerateKeyFramesExecutor) {
	fuegofw.Post(s.Manager, "/reelgen/keyframes",
		func(c fuegofw.ContextWithBody[in.GenerateKeyFramesRequest]) (in.GenerateKeyFramesResponse, error) {
			body, err := c.Body()
			if err != nil {
				return in.GenerateKeyFramesResponse{}, err
			}
			body.ProviderAPIKey = c.Header("x-api-key")
			return uc.Execute(c.Context(), body)
		})
}
