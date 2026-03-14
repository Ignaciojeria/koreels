package in

import "context"

// ConcatAudioRequest es el body de la generación de audio (beats con voice.audio.url por cada uno).
// voiceConfig se acepta y se devuelve en la respuesta para facilitar el pipeline (copiar/pegar entre pasos).
type ConcatAudioRequest struct {
	VoiceConfig *VoiceConfig `json:"voiceConfig,omitempty"`
	Beats       []Beat       `json:"beats"`
}

// ConcatAudioResponse devuelve el mismo body que el request con audio concatenado y sin URLs individuales por beat.
type ConcatAudioResponse struct {
	VoiceConfig *VoiceConfig    `json:"voiceConfig,omitempty"`
	Audio       ConcatAudioOutput `json:"audio"`
	Beats       []Beat            `json:"beats"` // mismos beats que el request pero con voice.audio sin URL
}

// ConcatAudioOutput agrupa voice, music y fullTrack.
type ConcatAudioOutput struct {
	Voice AudioTrack `json:"voice"`
}

// AudioTrack URL y duración de un track de audio.
type AudioTrack struct {
	URL      string  `json:"url,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

// MusicTrack URL y volumen (para cuando se agregue música).
type MusicTrack struct {
	URL    string  `json:"url,omitempty"`
	Volume float64 `json:"volume,omitempty"`
}

// ConcatAudioExecutor concatena los WAV de los beats y sube el resultado.
type ConcatAudioExecutor interface {
	Execute(ctx context.Context, req ConcatAudioRequest) (ConcatAudioResponse, error)
}
