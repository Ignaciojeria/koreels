package usecase

import (
	"context"
	"fmt"

	"koreels/internal/reelgen/application/ports/in"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewRunPipelineUseCase)

type runPipelineUseCase struct {
	beats  in.GenerateBeatsExecutor
	audio  in.GenerateAudioExecutor
	concat in.ConcatAudioExecutor
}

func NewRunPipelineUseCase(
	beats in.GenerateBeatsExecutor,
	audio in.GenerateAudioExecutor,
	concat in.ConcatAudioExecutor,
) in.RunPipelineExecutor {
	return &runPipelineUseCase{beats: beats, audio: audio, concat: concat}
}

func (u *runPipelineUseCase) Execute(ctx context.Context, req in.RunPipelineRequest) (in.ConcatAudioResponse, error) {
	beatsResp, err := u.beats.Execute(ctx, in.GenerateBeatsRequest{
		ScriptText:     req.ScriptText,
		LanguageCode:   req.LanguageCode,
		Subtitle:       req.Subtitle,
		ProviderAPIKey: req.ProviderAPIKey,
	})
	if err != nil {
		return in.ConcatAudioResponse{}, fmt.Errorf("generate beats: %w", err)
	}

	audioResp, err := u.audio.Execute(ctx, in.GenerateAudioRequest{
		VoiceConfig:    req.VoiceConfig,
		Beats:          beatsResp.Beats,
		ProviderAPIKey: req.ProviderAPIKey,
	})
	if err != nil {
		return in.ConcatAudioResponse{}, fmt.Errorf("generate audio: %w", err)
	}

	concatResp, err := u.concat.Execute(ctx, in.ConcatAudioRequest{
		VoiceConfig: req.VoiceConfig,
		Beats:       audioResp.Beats,
	})
	if err != nil {
		return in.ConcatAudioResponse{}, fmt.Errorf("concat audio: %w", err)
	}

	return concatResp, nil
}
