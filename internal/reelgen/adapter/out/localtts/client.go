package localtts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"koreels/internal/reelgen/adapter/out/ttscore"
	"koreels/internal/reelgen/application/ports/out"
	"koreels/internal/shared/configuration"
	"koreels/internal/shared/infrastructure/observability"

	"github.com/Ignaciojeria/ioc"
	"google.golang.org/genai"
)

var _ = ioc.Register(NewLocalTTSClient)

type localTTSClient struct {
	genaiClient *genai.Client
	outputDir   string
	obs         observability.Observability
}

func NewLocalTTSClient(
	genaiClient *genai.Client,
	conf configuration.Conf,
	obs observability.Observability,
) (out.TTSClient, error) {
	dir := conf.OUTPUT_DIR
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create output dir %q: %w", dir, err)
	}
	return &localTTSClient{
		genaiClient: genaiClient,
		outputDir:   dir,
		obs:         obs,
	}, nil
}

func (c *localTTSClient) GenerateSpeech(ctx context.Context, text string, apiKey string, opts *out.VoiceOptions) (*out.AudioResult, error) {
	client := c.genaiClient
	if client == nil && apiKey != "" {
		var err error
		client, err = genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  strings.TrimSpace(apiKey),
			Backend: genai.BackendGeminiAPI,
		})
		if err != nil {
			return nil, fmt.Errorf("genai client from api key: %w", err)
		}
	}
	if client == nil {
		return nil, fmt.Errorf("genai client is not initialized (set GEMINI_API_KEY or pass api key)")
	}
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text is required and cannot be empty")
	}

	lang := ttscore.DefaultLanguage
	voiceName := ttscore.DefaultVoice
	style := ""
	if opts != nil {
		if opts.Language != "" {
			lang = opts.Language
		}
		if opts.Voice != "" {
			voiceName = opts.Voice
		}
		style = opts.Style
	}

	var promptCtx *ttscore.PromptContext
	if opts != nil && opts.FullScript != "" {
		promptCtx = &ttscore.PromptContext{
			FullScript: opts.FullScript,
			BeatIndex:  opts.BeatIndex,
			TotalBeats: opts.TotalBeats,
			PrevText:   opts.PrevText,
		}
	}
	promptText := ttscore.BuildTTSPrompt(lang, style, text, promptCtx)

	pcmBytes, err := ttscore.SynthesizeToPCM(ctx, client, promptText, voiceName)
	if err != nil {
		c.obs.Logger.ErrorContext(ctx, "error_synthesizing_speech", "error", err)
		return nil, err
	}

	durationSeconds := float64(len(pcmBytes)) / float64(ttscore.BytesPerSecond)
	wavBytes := ttscore.PcmToWAV(pcmBytes, 24000, 1, 16)

	fileName := fmt.Sprintf("beat_%d.wav", time.Now().UnixNano())
	filePath := filepath.Join(c.outputDir, fileName)
	if err := os.WriteFile(filePath, wavBytes, 0o644); err != nil {
		return nil, fmt.Errorf("write wav to %s: %w", filePath, err)
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	c.obs.Logger.InfoContext(ctx, "reelgen_tts_audio_saved_local", "path", absPath, "duration_sec", durationSeconds)
	return &out.AudioResult{
		URL:      absPath,
		Duration: durationSeconds,
	}, nil
}
