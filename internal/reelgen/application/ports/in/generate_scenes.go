package in

import "context"

// GenerateScenesRequest es el body original: audio (voice) + beats. Opcionalmente visual_direction. Se hidrata con scenes en la respuesta.
type GenerateScenesRequest struct {
	Audio            *ConcatAudioOutput `json:"audio,omitempty"`
	VisualDirection  *VisualDirection  `json:"visual_direction,omitempty"`
	Beats            []Beat            `json:"beats"`
	ProviderAPIKey   string            `json:"-"` // Opcional; si viene por header, se usa para generar con LLM.
}

// GenerateScenesResponse devuelve el mismo body (audio + beats + visual_direction) hidratado con el plan de escenas.
type GenerateScenesResponse struct {
	Audio           *ConcatAudioOutput `json:"audio,omitempty"`
	VisualDirection *VisualDirection   `json:"visual_direction,omitempty"`
	Beats           []Beat             `json:"beats"`
	Scenes          []Scene           `json:"scenes"`
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

// Scene agrupa beats consecutivos con un prompt de video y opcional cámara/transición.
type Scene struct {
	BeatIDs       []int          `json:"beat_ids"`
	VideoPrompt   string         `json:"video_prompt"`              // Frase corta para el modelo de generación de video (solo texto).
	Camera        *CameraConfig  `json:"camera,omitempty"`          // Opcional: plano y movimiento.
	TransitionOut *TransitionOut `json:"transition_out,omitempty"`
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
