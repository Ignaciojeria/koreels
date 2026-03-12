package out

import "context"

type AudioResult struct {
	URL      string
	Duration float64
}

type TTSClient interface {
	GenerateSpeech(ctx context.Context, text string, apiKey string) (*AudioResult, error)
}
