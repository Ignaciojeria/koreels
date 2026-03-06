package qwenapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"koreels/internal/reelgen/application/ports/out"
	"koreels/internal/reelgen/domain/entity"
	"koreels/internal/shared/configuration"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewChatCompletionClient)

type chatCompletionClient struct {
	client *http.Client
	conf   configuration.Conf
}

// defaultTimeout para chat completions (LLM puede tardar).
const defaultTimeout = 600 * time.Second

func NewChatCompletionClient(conf configuration.Conf) out.ChatCompletionClient {
	return &chatCompletionClient{
		client: &http.Client{Timeout: defaultTimeout},
		conf:   conf,
	}
}

func (c *chatCompletionClient) Generate(ctx context.Context, systemPrompt, userPrompt string, responseFormat interface{}) (*entity.ChatCompletionResponse, error) {
	url := "https://dashscope-us.aliyuncs.com/compatible-mode/v1/chat/completions"

	requestBody := map[string]interface{}{
		"model": "qwen-plus",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": userPrompt,
			},
		},
	}

	if responseFormat != nil {
		requestBody["response_format"] = responseFormat
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
	apiKey := strings.TrimSpace(c.conf.DASHSCOPE_API_KEY)
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var responseData entity.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &responseData, nil
}
