package geminiapi

import (
	"context"
	"fmt"

	"koreels/internal/reelgen/domain/entity"

	"github.com/Ignaciojeria/ioc"
	"google.golang.org/genai"
)

var _ = ioc.Register(NewChatCompletionClient)

type ChatCompletionClient struct {
	client *genai.Client
}

func NewChatCompletionClient(client *genai.Client) *ChatCompletionClient {
	return &ChatCompletionClient{
		client: client,
	}
}

func (c *ChatCompletionClient) Generate(ctx context.Context, systemPrompt, userPrompt string, responseFormat interface{}, apiKey string) (*entity.ChatCompletionResponse, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("gemini api key is required")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gemini client: %w", err)
	}

	modelName := "gemini-2.5-pro"

	// Match the structured output for beats
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"beats": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"id": {Type: genai.TypeInteger},
						"voice": {
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"text": {Type: genai.TypeString},
							},
							Required: []string{"text"},
						},
						"subtitle": {
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"placement": {Type: genai.TypeString},
								"animation": {Type: genai.TypeString},
								"lines": {
									Type: genai.TypeArray,
									Items: &genai.Schema{
										Type: genai.TypeObject,
										Properties: map[string]*genai.Schema{
											"text": {Type: genai.TypeString},
											"emphasis": {
												Type:  genai.TypeArray,
												Items: &genai.Schema{Type: genai.TypeString},
											},
										},
										Required: []string{"text", "emphasis"},
									},
								},
							},
							Required: []string{"placement", "animation", "lines"},
						},
					},
					Required: []string{"id", "voice", "subtitle"},
				},
			},
		},
		Required: []string{"beats"},
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Parts: []*genai.Part{genai.NewPartFromText(systemPrompt)}},
		ResponseMIMEType:  "application/json",
		ResponseSchema:    schema,
	}

	contents := []*genai.Content{
		{Role: "user", Parts: []*genai.Part{genai.NewPartFromText(userPrompt)}},
	}

	resp, err := client.Models.GenerateContent(ctx, modelName, contents, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content from gemini: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates returned from gemini")
	}

	var jsonStr string
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			jsonStr = part.Text
			break
		}
	}

	if jsonStr == "" {
		return nil, fmt.Errorf("empty text in gemini response")
	}

	// We wrap it in the expected OpenAI-like entity structure so that the usecase doesn't change
	return &entity.ChatCompletionResponse{
		Choices: []entity.Choice{
			{
				Message: entity.Message{
					Role:    "assistant",
					Content: jsonStr,
				},
			},
		},
	}, nil
}
