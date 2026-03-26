package fuegoapi

import (
	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
)

var _ = ioc.Register(NewPostRunPipeline)

func NewPostRunPipeline(s *httpserver.Server, uc in.RunPipelineExecutor) {
	fuegofw.Post(s.Manager, "/reelgen/pipeline",
		func(c fuegofw.ContextWithBody[in.RunPipelineRequest]) (in.ConcatAudioResponse, error) {
			body, err := c.Body()
			if err != nil {
				return in.ConcatAudioResponse{}, err
			}
			body.ProviderAPIKey = c.Header("x-api-key")
			return uc.Execute(c.Context(), body)
		})
}
