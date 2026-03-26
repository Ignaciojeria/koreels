package out

import "context"

type AudioResult struct {
	URL      string
	Duration float64
}

// VoiceOptions opciones de voz para una llamada TTS (idioma, voz y estilo).
type VoiceOptions struct {
	Language string // BCP-47, ej. "es-ES", "en-US"
	Voice    string // nombre de voz, ej. "Sadachbia", "Kore"
	Style    string // instrucción de entrega, ej. "energetically and quickly"

	FullScript string // script completo para contexto narrativo
	BeatIndex  int    // posición del beat actual (0-based)
	TotalBeats int    // total de beats en el reel
	PrevText   string // texto del beat anterior para continuidad
}

type TTSClient interface {
	GenerateSpeech(ctx context.Context, text string, apiKey string, opts *VoiceOptions) (*AudioResult, error)
}
