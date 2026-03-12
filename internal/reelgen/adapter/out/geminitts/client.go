package geminitts

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"koreels/internal/reelgen/application/ports/out"
	"koreels/internal/shared/configuration"
	"koreels/internal/shared/infrastructure/observability"

	gcs "cloud.google.com/go/storage"
	"github.com/Ignaciojeria/ioc"
	"google.golang.org/genai"
)

const ttsModel = "gemini-2.5-flash-preview-tts"
const defaultVoice = "Sadachbia"
const bytesPerSecond = 24000 * 2 // PCM s16le 24kHz mono

var _ = ioc.Register(NewTTSClient)

type ttsClient struct {
	genaiClient *genai.Client
	gcsClient   *gcs.Client
	conf        configuration.Conf
	obs         observability.Observability
}

// NewTTSClient returns a TTSClient that uses Gemini TTS and uploads WAV to GCS, returning the public URL.
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

	lang := "es"
	voiceName := defaultVoice
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
	promptText := buildTTSPrompt(lang, style, text)

	pcmBytes, err := c.synthesizeToPCM(ctx, client, promptText, voiceName)
	if err != nil {
		return nil, err
	}

	durationSeconds := float64(len(pcmBytes)) / float64(bytesPerSecond)
	wavBytes := pcmToWAV(pcmBytes, 24000, 1, 16)

	if c.gcsClient == nil || c.conf.GCS_BUCKET == "" {
		// No GCS: return a placeholder URL (e.g. for local dev without GCS)
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

func (c *ttsClient) synthesizeToPCM(ctx context.Context, genaiClient *genai.Client, promptText string, voiceName string) ([]byte, error) {
	contents := []*genai.Content{
		{Role: "user", Parts: []*genai.Part{{Text: promptText}}},
	}
	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{string(genai.ModalityAudio)},
		SpeechConfig: &genai.SpeechConfig{
			VoiceConfig: &genai.VoiceConfig{
				PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
					VoiceName: voiceName,
				},
			},
		},
	}
	resp, err := genaiClient.Models.GenerateContent(ctx, ttsModel, contents, config)
	if err != nil {
		c.obs.Logger.ErrorContext(ctx, "error_synthesizing_speech", "error", err)
		return nil, fmt.Errorf("synthesizing speech: %w", err)
	}
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, fmt.Errorf("no audio generated (empty response)")
	}
	var pcmBytes []byte
	for _, p := range resp.Candidates[0].Content.Parts {
		if p != nil && p.InlineData != nil && len(p.InlineData.Data) > 0 {
			pcmBytes = p.InlineData.Data
			break
		}
	}
	if len(pcmBytes) == 0 {
		return nil, fmt.Errorf("synthesized audio is empty")
	}
	return pcmBytes, nil
}

func buildTTSPrompt(lang, style, text string) string {
	// Normalize "es-ES", "en-US" -> "es", "en" for the switch
	if i := strings.Index(lang, "-"); i > 0 {
		lang = lang[:i]
	}
	lang = strings.ToLower(strings.TrimSpace(lang))
	base := "Say in Spanish"
	switch lang {
	case "en":
		base = "Say in English"
	case "pt":
		base = "Say in Portuguese"
	case "fr":
		base = "Say in French"
	case "de":
		base = "Say in German"
	case "it":
		base = "Say in Italian"
	case "ja":
		base = "Say in Japanese"
	}
	if style != "" {
		style = strings.TrimSpace(style)
		style = strings.TrimSuffix(style, ":")
		return base + ", " + style + ": " + text
	}
	return base + " " + text
}

func pcmToWAV(pcm []byte, sampleRate, numChannels, bitsPerSample int) []byte {
	dataSize := len(pcm)
	byteRate := sampleRate * numChannels * (bitsPerSample / 8)
	blockAlign := numChannels * (bitsPerSample / 8)
	fileSize := 4 + 24 + 8 + dataSize
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, []byte("RIFF"))
	binary.Write(buf, binary.LittleEndian, uint32(fileSize))
	binary.Write(buf, binary.LittleEndian, []byte("WAVE"))
	binary.Write(buf, binary.LittleEndian, []byte("fmt "))
	binary.Write(buf, binary.LittleEndian, uint32(16))
	binary.Write(buf, binary.LittleEndian, uint16(1))
	binary.Write(buf, binary.LittleEndian, uint16(numChannels))
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(buf, binary.LittleEndian, uint32(byteRate))
	binary.Write(buf, binary.LittleEndian, uint16(blockAlign))
	binary.Write(buf, binary.LittleEndian, uint16(bitsPerSample))
	binary.Write(buf, binary.LittleEndian, []byte("data"))
	binary.Write(buf, binary.LittleEndian, uint32(dataSize))
	buf.Write(pcm)
	return buf.Bytes()
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
