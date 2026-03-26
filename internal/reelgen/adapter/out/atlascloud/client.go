package atlascloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"koreels/internal/reelgen/application/ports/out"
	"koreels/internal/shared/infrastructure/observability"

	"github.com/Ignaciojeria/ioc"
)

const (
	baseURL      = "https://api.atlascloud.ai/api/v1"
	model        = "vidu/q3-turbo/text-to-video"
	pollInterval = 10 * time.Second
	pollTimeout  = 6 * time.Minute
)

var _ = ioc.Register(NewVideoGenerationClient)

type videoClient struct {
	httpClient *http.Client
	logger     *slog.Logger
}

func NewVideoGenerationClient(obs observability.Observability) out.VideoGenerationClient {
	return &videoClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     obs.Logger,
	}
}

type generateVideoRequest struct {
	Model         string `json:"model"`
	AspectRatio   string `json:"aspect_ratio"`
	Duration      int    `json:"duration"`
	Prompt        string `json:"prompt"`
	Resolution    string `json:"resolution"`
	Style         string `json:"style"`
	GenerateAudio bool   `json:"generate_audio"`
}

type generateVideoResponse struct {
	Code int `json:"code"`
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

type predictionResponse struct {
	Code int `json:"code"`
	Data struct {
		ID      string   `json:"id"`
		Status  string   `json:"status"`
		Outputs []string `json:"outputs"`
		Error   string   `json:"error"`
	} `json:"data"`
}

func (c *videoClient) GenerateVideo(ctx context.Context, prompt string, aspectRatio string, duration int, apiKey string) (*out.VideoGenerationResult, error) {
	if duration < 1 {
		duration = 1
	}

	c.logger.InfoContext(ctx, "atlascloud_submit_video",
		"duration", duration,
		"aspect_ratio", aspectRatio,
		"prompt_preview", truncate(prompt, 80))

	predictionID, err := c.submitGeneration(ctx, prompt, aspectRatio, duration, apiKey)
	if err != nil {
		c.logger.ErrorContext(ctx, "atlascloud_submit_failed", "error", err)
		return nil, fmt.Errorf("submit: %w", err)
	}

	c.logger.InfoContext(ctx, "atlascloud_polling_started", "prediction_id", predictionID)

	videoURL, err := c.pollUntilComplete(ctx, predictionID, apiKey)
	if err != nil {
		c.logger.ErrorContext(ctx, "atlascloud_poll_failed", "prediction_id", predictionID, "error", err)
		return nil, fmt.Errorf("poll %s: %w", predictionID, err)
	}

	c.logger.InfoContext(ctx, "atlascloud_video_ready", "prediction_id", predictionID, "video_url", videoURL)
	return &out.VideoGenerationResult{VideoURL: videoURL}, nil
}

func (c *videoClient) submitGeneration(ctx context.Context, prompt string, aspectRatio string, duration int, apiKey string) (string, error) {
	reqBody := generateVideoRequest{
		Model:         model,
		AspectRatio:   aspectRatio,
		Duration:      duration,
		Prompt:        prompt,
		Resolution:    "720p",
		Style:         "general",
		GenerateAudio: false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/model/generateVideo", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var result generateVideoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if result.Data.ID == "" {
		return "", fmt.Errorf("empty prediction id in response: %s", string(body))
	}

	return result.Data.ID, nil
}

func (c *videoClient) pollUntilComplete(ctx context.Context, predictionID string, apiKey string) (string, error) {
	deadline := time.Now().Add(pollTimeout)

	for {
		if time.Now().After(deadline) {
			return "", fmt.Errorf("timed out after %v waiting for prediction %s", pollTimeout, predictionID)
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(pollInterval):
		}

		prediction, err := c.getPrediction(ctx, predictionID, apiKey)
		if err != nil {
			return "", err
		}

		c.logger.InfoContext(ctx, "atlascloud_poll_status",
			"prediction_id", predictionID,
			"status", prediction.Data.Status)

		switch prediction.Data.Status {
		case "completed", "succeeded":
			if len(prediction.Data.Outputs) == 0 || prediction.Data.Outputs[0] == "" {
				return "", fmt.Errorf("prediction %s completed but no output URL", predictionID)
			}
			return prediction.Data.Outputs[0], nil
		case "failed":
			errMsg := prediction.Data.Error
			if errMsg == "" {
				errMsg = "unknown error"
			}
			return "", fmt.Errorf("prediction %s failed: %s", predictionID, errMsg)
		}
	}
}

func (c *videoClient) getPrediction(ctx context.Context, predictionID string, apiKey string) (*predictionResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/model/prediction/"+predictionID, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var result predictionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &result, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
