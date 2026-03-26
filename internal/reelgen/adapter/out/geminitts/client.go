package geminitts

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"koreels/internal/reelgen/adapter/out/ttscore"
	"koreels/internal/reelgen/application/ports/out"
	"koreels/internal/shared/configuration"
	"koreels/internal/shared/infrastructure/observability"

	gcs "cloud.google.com/go/storage"
	"github.com/Ignaciojeria/ioc"
	"google.golang.org/genai"
)

var _ = ioc.Register(NewTTSClient)

type ttsClient struct {
	genaiClient *genai.Client
	gcsClient   *gcs.Client
	conf        configuration.Conf
	obs         observability.Observability
}

func NewTTSClient(
	genaiClient *genai.Client,
	gcsClient *gcs.Client,
	conf configuration.Conf,
	obs observability.Observability,
) (out.TTSClient, error) {
	return &ttsClient{
		genaiClient: genaiClient,
		gcsClient:   gcsClient,
		conf:        conf,
		obs:         obs,
	}, nil
}

func (c *ttsClient) GenerateSpeech(ctx context.Context, text string, apiKey string, opts *out.VoiceOptions) (*out.AudioResult, error) {
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
		return nil, fmt.Errorf("genai client is not initialized (set GEMINI_API_KEY or GOOGLE_PROJECT_ID)")
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

	if c.gcsClient == nil || c.conf.GCS_BUCKET == "" {
		return &out.AudioResult{
			URL:      "",
			Duration: durationSeconds,
		}, nil
	}

	publicURL, err := c.uploadToGCS(ctx, wavBytes)
	if err != nil {
		return nil, fmt.Errorf("upload audio to GCS: %w", err)
	}

	c.obs.Logger.InfoContext(ctx, "reelgen_tts_audio_uploaded", "url", publicURL, "duration_sec", durationSeconds)
	return &out.AudioResult{
		URL:      publicURL,
		Duration: durationSeconds,
	}, nil
}

func (c *ttsClient) uploadToGCS(ctx context.Context, wavBytes []byte) (publicURL string, err error) {
	bucketName := c.conf.GCS_BUCKET
	fileName := fmt.Sprintf("%d.wav", time.Now().UnixNano())
	objectPath := fmt.Sprintf("reelgen/audio/%s", fileName)

	uploadURL, err := c.gcsClient.Bucket(bucketName).SignedURL(objectPath, &gcs.SignedURLOptions{
		Method:      "PUT",
		Expires:     time.Now().Add(15 * time.Minute),
		ContentType: "audio/wav",
	})
	if err != nil {
		return "", fmt.Errorf("signed url: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(wavBytes))
	if err != nil {
		return "", fmt.Errorf("create upload request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "audio/wav")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("upload status %d (bucket=%q, check GCS_BUCKET exists and credentials have storage.objects.create): %s", resp.StatusCode, bucketName, string(body))
	}

	publicURL = fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectPath)
	return publicURL, nil
}
