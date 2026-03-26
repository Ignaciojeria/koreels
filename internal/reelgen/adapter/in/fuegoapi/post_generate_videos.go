package fuegoapi

import (
	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
)

var _ = ioc.Register(NewPostGenerateVideos)

func NewPostGenerateVideos(s *httpserver.Server, uc in.GenerateVideosExecutor) {
	fuegofw.Post(s.Manager, "/reelgen/videos",
		func(c fuegofw.ContextWithBody[in.GenerateVideosRequest]) (in.GenerateVideosResponse, error) {
			body, err := c.Body()
			if err != nil {
				return in.GenerateVideosResponse{}, err
			}
			body.ProviderAPIKey = c.Header("x-api-key")
			return uc.Execute(c.Context(), body)
		})
}
