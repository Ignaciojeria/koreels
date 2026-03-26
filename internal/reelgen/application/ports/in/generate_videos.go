package in

import "context"

// GenerateVideosRequest recibe beats con Prompt (de /keyframes) para generar un video por beat.
type GenerateVideosRequest struct {
	VoiceConfig     *VoiceConfig       `json:"voiceConfig,omitempty"`
	Audio           *ConcatAudioOutput `json:"audio,omitempty"`
	Beats           []Beat             `json:"beats"`
	VisualDirection *VisualDirection   `json:"visual_direction,omitempty"`
	ProviderAPIKey  string             `json:"-"`
}

// GenerateVideosResponse devuelve los beats con VideoURL llenado.
type GenerateVideosResponse struct {
	VoiceConfig     *VoiceConfig       `json:"voiceConfig,omitempty"`
	Audio           *ConcatAudioOutput `json:"audio,omitempty"`
	Beats           []Beat             `json:"beats"`
	VisualDirection *VisualDirection   `json:"visual_direction,omitempty"`
}

// GenerateVideosExecutor genera un video por beat secuencialmente.
type GenerateVideosExecutor interface {
	Execute(ctx context.Context, req GenerateVideosRequest) (GenerateVideosResponse, error)
}
