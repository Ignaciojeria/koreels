package usecase

import (
	"context"
	"fmt"
	"math"

	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/reelgen/application/ports/out"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewGenerateVideosUseCase)

type generateVideosUseCase struct {
	videoClient out.VideoGenerationClient
}

func NewGenerateVideosUseCase(videoClient out.VideoGenerationClient) in.GenerateVideosExecutor {
	return &generateVideosUseCase{videoClient: videoClient}
}

func (u *generateVideosUseCase) Execute(ctx context.Context, req in.GenerateVideosRequest) (in.GenerateVideosResponse, error) {
	if len(req.Beats) == 0 {
		return in.GenerateVideosResponse{}, fmt.Errorf("beats is required")
	}
	if req.ProviderAPIKey == "" {
		return in.GenerateVideosResponse{}, fmt.Errorf("x-api-key header is required")
	}

	beats := make([]in.Beat, len(req.Beats))
	copy(beats, req.Beats)

	aspectRatio := "9:16"
	if req.VisualDirection != nil && req.VisualDirection.AspectRatio != "" {
		aspectRatio = req.VisualDirection.AspectRatio
	}

	for i := range beats {
		if beats[i].Prompt == "" || beats[i].VideoURL != "" {
			continue
		}

		duration := beatVideoDuration(beats[i])
		result, err := u.videoClient.GenerateVideo(ctx, beats[i].Prompt, aspectRatio, duration, req.ProviderAPIKey)
		if err != nil {
			return in.GenerateVideosResponse{
				VoiceConfig:     req.VoiceConfig,
				Audio:           req.Audio,
				Beats:           beats,
				VisualDirection: req.VisualDirection,
			}, fmt.Errorf("beat %d: %w", beats[i].ID, err)
		}
		beats[i].VideoURL = result.VideoURL
		break
	}

	return in.GenerateVideosResponse{
		VoiceConfig:     req.VoiceConfig,
		Audio:           req.Audio,
		Beats:           beats,
		VisualDirection: req.VisualDirection,
	}, nil
}

func beatVideoDuration(b in.Beat) int {
	if b.Voice.Audio == nil || b.Voice.Audio.Duration <= 0 {
		return 5
	}
	d := int(math.Ceil(b.Voice.Audio.Duration))
	if d < 1 {
		d = 1
	}
	return d
}
