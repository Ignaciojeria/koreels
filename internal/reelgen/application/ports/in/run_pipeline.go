package in

import "context"

// RunPipelineRequest combina los inputs de beats, audio y concat en una sola llamada.
type RunPipelineRequest struct {
	ScriptText     string         `json:"scriptText"`
	LanguageCode   string         `json:"languageCode"`
	VoiceConfig    *VoiceConfig   `json:"voiceConfig,omitempty"`
	Subtitle       SubtitleConfig `json:"subtitle"`
	ProviderAPIKey string         `json:"-"`
}

// RunPipelineExecutor orquestra beats -> audio -> concat en secuencia.
// Retorna ConcatAudioResponse (listo para pasar a /keyframes).
type RunPipelineExecutor interface {
	Execute(ctx context.Context, req RunPipelineRequest) (ConcatAudioResponse, error)
}
