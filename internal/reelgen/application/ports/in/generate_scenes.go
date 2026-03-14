package in

import "context"

// GenerateScenesRequest es el body original: audio (voice) + beats. Opcionalmente visual_direction y voiceConfig (para pipeline).
type GenerateScenesRequest struct {
	VoiceConfig     *VoiceConfig       `json:"voiceConfig,omitempty"`
	Audio           *ConcatAudioOutput `json:"audio,omitempty"`
	VisualDirection *VisualDirection   `json:"visual_direction,omitempty"`
	Beats           []Beat             `json:"beats"`
	ProviderAPIKey  string             `json:"-"` // Opcional; si viene por header, se usa para generar con LLM.
}

// GenerateScenesResponse devuelve el mismo body (voiceConfig + audio + beats + visual_direction) hidratado con el plan de escenas.
type GenerateScenesResponse struct {
	VoiceConfig     *VoiceConfig       `json:"voiceConfig,omitempty"`
	Audio           *ConcatAudioOutput `json:"audio,omitempty"`
	VisualDirection *VisualDirection   `json:"visual_direction,omitempty"`
	Beats           []Beat             `json:"beats"`
	Scenes          []Scene            `json:"scenes"`
}

// VisualDirection estilo global para todo el reel (consistencia visual).
type VisualDirection struct {
	Style        string   `json:"style,omitempty"`         // ej. "cinematic tech"
	Environment  string   `json:"environment,omitempty"`   // ej. "modern dark workspace"
	Lighting     string   `json:"lighting,omitempty"`       // ej. "dramatic rim lighting, high contrast"
	ColorPalette []string `json:"color_palette,omitempty"`  // ej. ["deep red", "dark gray"]
	CameraStyle  string   `json:"camera_style,omitempty"`   // ej. "dynamic social media reel"
	AspectRatio  string   `json:"aspect_ratio,omitempty"`  // ej. "9:16"
}

// Scene agrupa beats consecutivos con prompt de video, duración calculada y slot para el asset generado.
// Listo para Remotion: duration a la vista; asset_url se hidrata cuando el video esté listo (async).
type Scene struct {
	ID             int            `json:"id"`                        // ID de escena para tracking y paralelización
	BeatIDs        []int          `json:"beat_ids"`
	Duration       float64        `json:"duration"`                 // Suma de duraciones de los beats (evita que el renderer recalcule)
	VideoPrompt    string         `json:"video_prompt"`             // Frase para el modelo de generación de video
	AssetURL       string         `json:"asset_url,omitempty"`      // Se llena al hidratar con el video generado (proceso async)
	Camera         *CameraConfig  `json:"camera,omitempty"`
	TransitionOut  *TransitionOut `json:"transition_out,omitempty"`
}

// CameraConfig sugiere plano y movimiento para la escena.
type CameraConfig struct {
	Shot     string `json:"shot,omitempty"`     // ej. "medium close-up", "screen close-up"
	Movement string `json:"movement,omitempty"` // ej. "slow push-in", "subtle parallax"
}

// TransitionOut tipo y duración de la transición al salir de la escena.
type TransitionOut struct {
	Type     string  `json:"type"`
	Duration float64 `json:"duration,omitempty"`
}

// GenerateScenesExecutor agrupa beats en escenas ≥ 4s y opcionalmente propone prompts visuales.
type GenerateScenesExecutor interface {
	Execute(ctx context.Context, req GenerateScenesRequest) (GenerateScenesResponse, error)
}
