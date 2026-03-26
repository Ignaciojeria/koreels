package out

import "context"

// VideoGenerationResult contiene la URL del video generado.
type VideoGenerationResult struct {
	VideoURL string
}

// VideoGenerationClient abstrae la generacion de video (submit + polling + resultado).
type VideoGenerationClient interface {
	GenerateVideo(ctx context.Context, prompt string, aspectRatio string, duration int, apiKey string) (*VideoGenerationResult, error)
}
