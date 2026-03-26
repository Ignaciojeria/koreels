package usecase

import (
	"context"
	"fmt"
	"strings"

	"koreels/internal/reelgen/application/ports/in"

	"github.com/Ignaciojeria/ioc"
)

const MinSceneDurationSeconds = 4.0

var _ = ioc.Register(NewGenerateScenesUseCase)

type generateScenesUseCase struct{}

func NewGenerateScenesUseCase() in.GenerateScenesExecutor {
	return &generateScenesUseCase{}
}

func (u *generateScenesUseCase) Execute(_ context.Context, req in.GenerateScenesRequest) (in.GenerateScenesResponse, error) {
	if err := validateBeats(req.Beats); err != nil {
		return in.GenerateScenesResponse{}, err
	}

	var scenes []in.Scene
	switch req.Grouping {
	case "beats":
		scenes = groupByBeat(req.Beats)
	default:
		scenes = groupByTiming(req.Beats)
	}

	return in.GenerateScenesResponse{
		VoiceConfig: req.VoiceConfig,
		Audio:       req.Audio,
		Beats:       req.Beats,
		Scenes:      scenes,
	}, nil
}

func groupByTiming(beats []in.Beat) []in.Scene {
	groups := groupBeatsIntoScenes(beats)
	byID := beatsByID(beats)
	scenes := make([]in.Scene, 0, len(groups))
	for i, beatIDs := range groups {
		scenes = append(scenes, in.Scene{
			ID:        i + 1,
			BeatIDs:   beatIDs,
			Duration:  sumDuration(byID, beatIDs),
			Narration: buildNarration(byID, beatIDs),
			KeyFrames: buildKeyFrames(byID, beatIDs),
		})
	}
	return scenes
}

func groupByBeat(beats []in.Beat) []in.Scene {
	scenes := make([]in.Scene, 0, len(beats))
	for i, b := range beats {
		duration := float64(0)
		if b.Voice.Audio != nil {
			duration = b.Voice.Audio.Duration
		}
		scenes = append(scenes, in.Scene{
			ID:        i + 1,
			BeatIDs:   []int{b.ID},
			Duration:  duration,
			Narration: b.Voice.Text,
			KeyFrames: buildKeyFrames(map[int]in.Beat{b.ID: b}, []int{b.ID}),
		})
	}
	return scenes
}

func validateBeats(beats []in.Beat) error {
	seen := make(map[int]bool)
	for i, b := range beats {
		if seen[b.ID] {
			return fmt.Errorf("beat id duplicado: %d", b.ID)
		}
		seen[b.ID] = true
		if i > 0 && beats[i-1].ID >= b.ID {
			return fmt.Errorf("beats deben estar en orden secuencial por id; encontrado id %d después de %d", b.ID, beats[i-1].ID)
		}
	}
	return nil
}

// groupBeatsIntoScenes agrupa beats consecutivos de modo que cada escena sume >= MinSceneDurationSeconds.
// Si la última escena quedaría < 4s, se fusiona con la anterior.
func groupBeatsIntoScenes(beats []in.Beat) [][]int {
	if len(beats) == 0 {
		return nil
	}

	var groups [][]int
	var current []int
	var duration float64

	for _, b := range beats {
		d := float64(0)
		if b.Voice.Audio != nil {
			d = b.Voice.Audio.Duration
		}
		current = append(current, b.ID)
		duration += d
		if duration >= MinSceneDurationSeconds {
			groups = append(groups, current)
			current = nil
			duration = 0
		}
	}

	if len(current) > 0 {
		if len(groups) > 0 && duration < MinSceneDurationSeconds {
			groups[len(groups)-1] = append(groups[len(groups)-1], current...)
		} else {
			groups = append(groups, current)
		}
	}

	return groups
}

func beatsByID(beats []in.Beat) map[int]in.Beat {
	m := make(map[int]in.Beat, len(beats))
	for _, b := range beats {
		m[b.ID] = b
	}
	return m
}

func sumDuration(byID map[int]in.Beat, beatIDs []int) float64 {
	var d float64
	for _, id := range beatIDs {
		if b, ok := byID[id]; ok && b.Voice.Audio != nil {
			d += b.Voice.Audio.Duration
		}
	}
	return d
}

func buildKeyFrames(byID map[int]in.Beat, beatIDs []int) []in.KeyFrame {
	var kfs []in.KeyFrame
	for _, id := range beatIDs {
		b, ok := byID[id]
		if !ok {
			continue
		}
		beatDuration := float64(0)
		if b.Voice.Audio != nil {
			beatDuration = b.Voice.Audio.Duration
		}

		subs := b.Subtitle.Lines
		if len(subs) == 0 {
			if b.Voice.Text != "" {
				kfs = append(kfs, in.KeyFrame{Narration: b.Voice.Text, Duration: beatDuration})
			}
			continue
		}

		totalChars := 0
		for _, l := range subs {
			totalChars += len([]rune(l.Text))
		}

		for _, l := range subs {
			lineChars := len([]rune(l.Text))
			lineDuration := beatDuration
			if totalChars > 0 {
				lineDuration = beatDuration * float64(lineChars) / float64(totalChars)
			}
			kfs = append(kfs, in.KeyFrame{Narration: l.Text, Duration: lineDuration})
		}
	}
	return kfs
}

func buildNarration(byID map[int]in.Beat, beatIDs []int) string {
	parts := make([]string, 0, len(beatIDs))
	for _, id := range beatIDs {
		if b, ok := byID[id]; ok && b.Voice.Text != "" {
			parts = append(parts, b.Voice.Text)
		}
	}
	return strings.Join(parts, " ")
}
