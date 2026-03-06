package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/reelgen/application/ports/out"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewGenerateScenesUseCase)

type generateScenesUseCase struct {
	client out.ChatCompletionClient
}

func NewGenerateScenesUseCase(client out.ChatCompletionClient) in.GenerateScenesExecutor {
	return &generateScenesUseCase{
		client: client,
	}
}

func (u *generateScenesUseCase) Execute(ctx context.Context, req in.GenerateScenesRequest) (in.GenerateScenesResponse, error) {
	systemPrompt := `You are an expert script segmentation assistant for short-form video content (like Reels/TikTok). Segment the script into coherent scenes following these rules:
1. Subtitle Timing: Use relative timestamps per scene (start at 0.0). Lines must be progressive, non-overlapping, and continuously cover the ENTIRE scene duration with zero time gaps.
2. Reading Pace: Target 2.5-3.5 words/sec. Give long sentences adequate time and split them. Never cram text (e.g. 15 words in 0.7s).
3. Quality: Omit empty lines. Only output real text.
4. Animations: 'SCALE_IN' for scene 0, 'SLIDE_UP' for the rest.
5. Placement: If 'placementStrategy' is TOP/BOTTOM, apply strictly to all scenes. If DYNAMIC, intelligently choose TOP, CENTER, or BOTTOM per scene.
6. Emphasis: Pick 1-2 high-impact words per line (no filler words) existing in the text and add them to the 'emphasis' array.
Respect user constraints (style, max chars/lines) and strictly output the requested JSON schema.`

	userPromptBytes, _ := json.Marshal(req)
	userPrompt := fmt.Sprintf("Please process the following request parameters and script:\n%s", string(userPromptBytes))

	responseFormat := map[string]interface{}{
		"type": "json_schema",
		"json_schema": map[string]interface{}{
			"name":   "generate_scenes_response",
			"strict": true,
			"schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scenes": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"index": map[string]interface{}{"type": "integer"},
								"start": map[string]interface{}{"type": "number"},
								"end":   map[string]interface{}{"type": "number"},
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
													"text":  map[string]interface{}{"type": "string"},
													"start": map[string]interface{}{"type": "number"},
													"end":   map[string]interface{}{"type": "number"},
													"emphasis": map[string]interface{}{
														"type":  "array",
														"items": map[string]interface{}{"type": "string"},
													},
												},
												"required":             []string{"text", "start", "end"},
												"additionalProperties": false,
											},
										},
									},
									"required":             []string{"placement", "animation", "lines"},
									"additionalProperties": false,
								},
							},
							"required":             []string{"index", "start", "end", "voice", "subtitle"},
							"additionalProperties": false,
						},
					},
				},
				"required":             []string{"scenes"},
				"additionalProperties": false,
			},
		},
	}

	completionResp, err := u.client.Generate(ctx, systemPrompt, userPrompt, responseFormat)
	if err != nil {
		return in.GenerateScenesResponse{}, fmt.Errorf("failed to generate scenes from qwen: %w", err)
	}

	if len(completionResp.Choices) == 0 {
		return in.GenerateScenesResponse{}, fmt.Errorf("no choices returned from qwen")
	}

	content := completionResp.Choices[0].Message.Content

	var resp in.GenerateScenesResponse
	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		return in.GenerateScenesResponse{}, fmt.Errorf("failed to unmarshal structured output: %w", err)
	}

	return resp, nil
}
