package in

import "context"

// VoiceConfig opciones de voz para TTS (idioma, voz y estilo de entrega).
type VoiceConfig struct {
	// Language: BCP-47. Por defecto el backend usa español latino (es-MX).
	// Enviar "es-ES" para español de España; "es-MX", "es-AR", etc. para variantes latinas.
	Language string `json:"language"` // ej. "es-MX" (default), "es-ES", "en-US"
	Voice    string `json:"voice"`    // nombre de voz TTS, ej. "Kore", "Puck"
	Style    string `json:"style"`   // instrucción de entrega, ej. "fast and energetically", "cheerfully"
}

type GenerateAudioRequest struct {
	VoiceConfig     *VoiceConfig   `json:"voiceConfig,omitempty"`
	Beats           []Beat         `json:"beats"`
	PartialFailures []BeatFailure  `json:"partial_failures,omitempty"` // ignorado; aceptado para poder pegar el response como request y reprocesar
	ProviderAPIKey  string         `json:"-"`                           // header x-api-key
}

// GenerateAudioResponse tiene la misma forma que el request (voiceConfig + beats) para copiar/pegar y reprocesar.
// Los beats incluyen las URLs de audio generadas; los que fallaron quedan sin audio. Opcional partial_failures.
type GenerateAudioResponse struct {
	VoiceConfig     *VoiceConfig   `json:"voiceConfig,omitempty"`
	Beats           []Beat         `json:"beats"`
	PartialFailures []BeatFailure  `json:"partial_failures,omitempty"` // beats que fallaron; copiar el body y reenviar para reintentar solo esos
}

// BeatFailure indica que no se pudo generar audio para ese beat (reintento solo ese).
type BeatFailure struct {
	BeatID int    `json:"beat_id"`
	Error  string `json:"error"`
}

type GenerateAudioExecutor interface {
	Execute(ctx context.Context, req GenerateAudioRequest) (GenerateAudioResponse, error)
}
