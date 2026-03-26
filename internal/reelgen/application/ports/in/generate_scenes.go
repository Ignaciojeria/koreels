package in

import "context"

// GenerateScenesRequest recibe beats con duraciones de audio para agruparlos en escenas.
// VoiceConfig y Audio se aceptan como pass-through del pipeline (no se usan, solo se devuelven).
type GenerateScenesRequest struct {
	VoiceConfig *VoiceConfig       `json:"voiceConfig,omitempty"`
	Audio       *ConcatAudioOutput `json:"audio,omitempty"`
	Beats       []Beat             `json:"beats"`
	Grouping    string             `json:"grouping,omitempty"` // "timing" (default) | "beats"
}

// GenerateScenesResponse devuelve el contexto completo del pipeline hidratado con las escenas.
type GenerateScenesResponse struct {
	VoiceConfig *VoiceConfig       `json:"voiceConfig,omitempty"`
	Audio       *ConcatAudioOutput `json:"audio,omitempty"`
	Beats       []Beat             `json:"beats"`
	Scenes      []Scene            `json:"scenes"`
}

// Scene agrupa beats consecutivos con keyframes estructurales por línea de subtítulo.
type Scene struct {
	ID        int        `json:"id"`
	BeatIDs   []int      `json:"beat_ids"`
	Duration  float64    `json:"duration"`
	Narration string     `json:"narration"`
	KeyFrames []KeyFrame `json:"keyframes"`
}

// KeyFrame representa un segmento visual dentro de una escena.
// Duration y Narration se llenan de forma determinista (scenes endpoint).
// Prompt y NegativePrompt se llenan via LLM (keyframes endpoint).
type KeyFrame struct {
	Prompt         string  `json:"prompt,omitempty"`
	NegativePrompt string  `json:"negative_prompt,omitempty"`
	Duration       float64 `json:"duration"`
	Narration      string  `json:"narration"`
}

// GenerateScenesExecutor agrupa beats en escenas de forma determinista (sin LLM).
type GenerateScenesExecutor interface {
	Execute(ctx context.Context, req GenerateScenesRequest) (GenerateScenesResponse, error)
}
