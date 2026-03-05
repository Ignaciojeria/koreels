package qwenapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Ignaciojeria/ioc"
	"koreels/internal/reelgen/application/ports/out"
	"koreels/internal/reelgen/domain/entity"
)

var _ = ioc.Register(NewChatCompletionClient)

type chatCompletionClient struct {
	client *http.Client
}

func NewChatCompletionClient() out.ChatCompletionClient {
	return &chatCompletionClient{
		client: &http.Client{},
	}
}

func (c *chatCompletionClient) Generate(ctx context.Context, prompt string) (*entity.ChatCompletionResponse, error) {
	// For now, this is a placeholder implementation expecting an OpenAI/Qwen compatible endpoint.
	// We'll construct a mock request just to satisfy the struct mapping as requested.

	url := "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions" // Example URL for Qwen

	requestBody := map[string]interface{}{
		"model": "qwen-plus",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Authorization", "Bearer YOUR_API_KEY")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var responseData entity.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &responseData, nil
}
