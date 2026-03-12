package fuegoapi

import (
	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
)

var _ = ioc.Register(NewPostConcatAudio)

// NewPostConcatAudio registra POST /reelgen/audio/concat: concatena los WAV de los beats y devuelve audio.voice con la URL del track completo.
func NewPostConcatAudio(s *httpserver.Server, uc in.ConcatAudioExecutor) {
	fuegofw.Post(s.Manager, "/reelgen/audio/concat",
		func(c fuegofw.ContextWithBody[in.ConcatAudioRequest]) (in.ConcatAudioResponse, error) {
			body, err := c.Body()
			if err != nil {
				return in.ConcatAudioResponse{}, err
			}
			return uc.Execute(c.Context(), body)
		})
}
