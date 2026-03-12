package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/reelgen/application/ports/out"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewGenerateBeatsUseCase)

type generateBeatsUseCase struct {
	client out.ChatCompletionClient
}

func NewGenerateBeatsUseCase(client out.ChatCompletionClient) in.GenerateBeatsExecutor {
	return &generateBeatsUseCase{
		client: client,
	}
}

func (u *generateBeatsUseCase) Execute(ctx context.Context, req in.GenerateBeatsRequest) (in.GenerateBeatsResponse, error) {
	systemPrompt := `You are an expert script segmentation assistant for short-form video content (like Reels/TikTok). You receive a text and instructions. You must break the text into smaller 'beats'.
Output MUST be JSON with a "beats" array.

RULES:
1. Beats: Subtitle text and voice segments. One beat per spoken phrase; split long sentences into multiple beats. Do not include start/end times.
2. ID: Beat "id" starts at 1 and increments.
3. Subtitle Lines: Split the text into subtitle lines according to user constraints (e.g., maxCharsPerLine, maxLines). Make sure it's readable.
4. Animations (Beats Subtitles): 'SCALE_IN' for the first beat, 'SLIDE_UP' for the rest.
5. Placement: Apply the requested placementStrategy. If DYNAMIC, choose TOP, CENTER, or BOTTOM per beat.
6. Emphasis: Pick 1-2 high-impact words per line (no filler words) existing in the text and add them to the 'emphasis' array.
Respect user constraints and strictly output the requested JSON schema.`

	userPromptBytes, _ := json.Marshal(req)
	userPrompt := fmt.Sprintf("Please process the following request parameters and script:\n%s", string(userPromptBytes))

	responseFormat := map[string]interface{}{
		"type": "json_schema",
		"json_schema": map[string]interface{}{
			"name":   "generate_beats_response",
			"strict": true,
			"schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"beats": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"id": map[string]interface{}{"type": "integer"},
								"voice": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"text": map[string]interface{}{"type": "string"},
									},
									"required":             []string{"text"},
									"additionalProperties": false,
								},
								"subtitle": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"placement": map[string]interface{}{"type": "string"},
										"animation": map[string]interface{}{"type": "string"},
										"lines": map[string]interface{}{
											"type": "array",
											"items": map[string]interface{}{
												"type": "object",
												"properties": map[string]interface{}{
													"text": map[string]interface{}{"type": "string"},
													"emphasis": map[string]interface{}{
														"type": "array",
														"items": map[string]interface{}{
															"type": "string",
														},
													},
												},
												"required":             []string{"text", "emphasis"},
												"additionalProperties": false,
											},
										},
									},
									"required":             []string{"placement", "animation", "lines"},
									"additionalProperties": false,
								},
							},
							"required":             []string{"id", "voice", "subtitle"},
							"additionalProperties": false,
						},
					},
				},
				"required":             []string{"beats"},
				"additionalProperties": false,
			},
		},
	}

	completionResp, err := u.client.Generate(ctx, systemPrompt, userPrompt, responseFormat, req.ProviderAPIKey)
	if err != nil {
		return in.GenerateBeatsResponse{}, fmt.Errorf("failed to generate beats from qwen: %w", err)
	}

	if len(completionResp.Choices) == 0 {
		return in.GenerateBeatsResponse{}, fmt.Errorf("no choices returned from qwen")
	}

	content := completionResp.Choices[0].Message.Content

	var resp in.GenerateBeatsResponse
	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		return in.GenerateBeatsResponse{}, fmt.Errorf("failed to unmarshal structured output: %w", err)
	}

	return resp, nil
}
