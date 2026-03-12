package in

import "context"

// VoiceConfig opciones de voz para TTS (idioma, voz y estilo de entrega).
type VoiceConfig struct {
	Language string `json:"language"` // BCP-47, ej. "es-ES", "en-US"
	Voice    string `json:"voice"`   // nombre de voz del proveedor TTS, ej. "Sadachbia", "Kore"
	Style    string `json:"style"`   // instrucción de entrega, ej. "energetically and quickly", "cheerfully", "at a brisk pace"
}

type GenerateAudioRequest struct {
	VoiceConfig     *VoiceConfig `json:"voiceConfig,omitempty"`
	Beats           []Beat       `json:"beats"`
	ProviderAPIKey  string       `json:"-"` // Just in case it's needed via header
}

type GenerateAudioResponse struct {
	Beats []Beat `json:"beats"`
}

type GenerateAudioExecutor interface {
	Execute(ctx context.Context, req GenerateAudioRequest) (GenerateAudioResponse, error)
}
