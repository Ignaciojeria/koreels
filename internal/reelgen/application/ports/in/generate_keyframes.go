package in

import "context"

// GenerateKeyFramesRequest recibe beats del pipeline y genera prompts visuales via LLM.
// Si visual_direction no se provee, se auto-genera via LLM.
type GenerateKeyFramesRequest struct {
	VoiceConfig     *VoiceConfig       `json:"voiceConfig,omitempty"`
	Audio           *ConcatAudioOutput `json:"audio,omitempty"`
	Beats           []Beat             `json:"beats"`
	VisualDirection *VisualDirection   `json:"visual_direction,omitempty"`
	ProviderAPIKey  string             `json:"-"`
}

// GenerateKeyFramesResponse devuelve beats con Prompt llenado por LLM.
type GenerateKeyFramesResponse struct {
	VoiceConfig     *VoiceConfig       `json:"voiceConfig,omitempty"`
	Audio           *ConcatAudioOutput `json:"audio,omitempty"`
	Beats           []Beat             `json:"beats"`
	VisualDirection *VisualDirection   `json:"visual_direction"`
}

// VisualDirection estilo visual global para todo el reel (consistencia entre beats).
type VisualDirection struct {
	Style        string   `json:"style,omitempty"`
	Environment  string   `json:"environment,omitempty"`
	Lighting     string   `json:"lighting,omitempty"`
	ColorPalette []string `json:"color_palette,omitempty"`
	CameraStyle  string   `json:"camera_style,omitempty"`
	AspectRatio  string   `json:"aspect_ratio,omitempty"`
}

// GenerateKeyFramesExecutor genera prompts visuales por beat via LLM (1-2 llamadas).
type GenerateKeyFramesExecutor interface {
	Execute(ctx context.Context, req GenerateKeyFramesRequest) (GenerateKeyFramesResponse, error)
}
