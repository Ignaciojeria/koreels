package usecase

import (
	"context"
	"fmt"

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
		Beats: req.Beats,
	}

	for i := range resp.Beats {
		beat := &resp.Beats[i]
		if beat.Voice.Text == "" {
			continue
		}
		res, err := u.client.GenerateSpeech(ctx, beat.Voice.Text, req.ProviderAPIKey)
		if err != nil {
			return in.GenerateAudioResponse{}, fmt.Errorf("beat %d: %w", beat.ID, err)
		}
		beat.Voice.Audio = &in.Audio{
			URL:      res.URL,
			Duration: res.Duration,
		}
	}

	return resp, nil
}
