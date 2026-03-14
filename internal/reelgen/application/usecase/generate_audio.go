package usecase

import (
	"context"

	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/reelgen/application/ports/out"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewGenerateAudioUseCase)

type generateAudioUseCase struct {
	client out.TTSClient
}

func NewGenerateAudioUseCase(client out.TTSClient) in.GenerateAudioExecutor {
	return &generateAudioUseCase{
		client: client,
	}
}

func (u *generateAudioUseCase) Execute(ctx context.Context, req in.GenerateAudioRequest) (in.GenerateAudioResponse, error) {
	resp := in.GenerateAudioResponse{
		VoiceConfig: req.VoiceConfig,
		Beats:       req.Beats,
	}

	var voiceOpts *out.VoiceOptions
	if req.VoiceConfig != nil {
		voiceOpts = &out.VoiceOptions{
			Language: req.VoiceConfig.Language,
			Voice:    req.VoiceConfig.Voice,
			Style:    req.VoiceConfig.Style,
		}
	}

	var partialFailures []in.BeatFailure
	for i := range resp.Beats {
		beat := &resp.Beats[i]
		if beat.Voice.Text == "" {
			continue
		}
		// Reintento: si el beat ya tiene audio URL, no volver a generar (hidratar con lo que ya vino).
		if beat.Voice.Audio != nil && beat.Voice.Audio.URL != "" {
			continue
		}
		res, err := u.client.GenerateSpeech(ctx, beat.Voice.Text, req.ProviderAPIKey, voiceOpts)
		if err != nil {
			partialFailures = append(partialFailures, in.BeatFailure{BeatID: beat.ID, Error: err.Error()})
			continue
		}
		beat.Voice.Audio = &in.Audio{
			URL:      res.URL,
			Duration: res.Duration,
		}
	}

	resp.PartialFailures = partialFailures
	return resp, nil
}
