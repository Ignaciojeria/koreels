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

	var partialFailures []in.BeatFailure
	for i := range resp.Beats {
		beat := &resp.Beats[i]
		if beat.Voice.Text == "" {
			continue
		}
		if beat.Voice.Audio != nil && beat.Voice.Audio.URL != "" {
			continue
		}

		voiceOpts := &out.VoiceOptions{
			BeatID: beat.ID,
		}
		if req.VoiceConfig != nil {
			voiceOpts.Language = req.VoiceConfig.Language
			voiceOpts.Voice = req.VoiceConfig.Voice
			voiceOpts.Style = req.VoiceConfig.Style
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
