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
	"koreels/internal/shared/configuration"
)

var _ = ioc.Register(NewChatCompletionClient)

type chatCompletionClient struct {
	client *http.Client
	conf   configuration.Conf
}

func NewChatCompletionClient(conf configuration.Conf) out.ChatCompletionClient {
	return &chatCompletionClient{
		client: &http.Client{},
		conf:   conf,
	}
}

func (c *chatCompletionClient) Generate(ctx context.Context, prompt string) (*entity.ChatCompletionResponse, error) {
	url := "https://dashscope-intl.aliyuncs.com/compatible-mode/v1/chat/completions"

	requestBody := map[string]interface{}{
		"model": "qwen-plus",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a helpful assistant.",
			},
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
	if c.conf.DASHSCOPE_API_KEY != "" {
		req.Header.Set("Authorization", "Bearer "+c.conf.DASHSCOPE_API_KEY)
	}

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
