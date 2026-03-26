package usecase

import (
	"context"
	"strings"

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

	fullScript := buildFullScript(resp.Beats)
	totalBeats := countSpeakableBeats(resp.Beats)

	var partialFailures []in.BeatFailure
	var prevText string
	speakableIndex := 0
	for i := range resp.Beats {
		beat := &resp.Beats[i]
		if beat.Voice.Text == "" {
			continue
		}
		if beat.Voice.Audio != nil && beat.Voice.Audio.URL != "" {
			prevText = beat.Voice.Text
			speakableIndex++
			continue
		}

		voiceOpts := &out.VoiceOptions{
			FullScript: fullScript,
			BeatIndex:  speakableIndex,
			TotalBeats: totalBeats,
			PrevText:   prevText,
		}
		if req.VoiceConfig != nil {
			voiceOpts.Language = req.VoiceConfig.Language
			voiceOpts.Voice = req.VoiceConfig.Voice
			voiceOpts.Style = req.VoiceConfig.Style
		}

		res, err := u.client.GenerateSpeech(ctx, beat.Voice.Text, req.ProviderAPIKey, voiceOpts)
		if err != nil {
			partialFailures = append(partialFailures, in.BeatFailure{BeatID: beat.ID, Error: err.Error()})
			speakableIndex++
			continue
		}
		beat.Voice.Audio = &in.Audio{
			URL:      res.URL,
			Duration: res.Duration,
		}
		prevText = beat.Voice.Text
		speakableIndex++
	}

	resp.PartialFailures = partialFailures
	return resp, nil
}

func buildFullScript(beats []in.Beat) string {
	var parts []string
	for _, b := range beats {
		if b.Voice.Text != "" {
			parts = append(parts, b.Voice.Text)
		}
	}
	return strings.Join(parts, " ")
}

func countSpeakableBeats(beats []in.Beat) int {
	n := 0
	for _, b := range beats {
		if b.Voice.Text != "" {
			n++
		}
	}
	return n
}
