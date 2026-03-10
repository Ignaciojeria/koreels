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
	systemPrompt := `You are an expert script segmentation and video direction assistant for short-form video content (like Reels/TikTok). You receive a text and instructions. You must separate the visual timeline from the text timeline.
Output MUST be JSON with "duration", "visualDirection", "scenes", and "beats".

RULES:
1. Duration: Total estimated duration of the short video.
2. VisualDirection: Use the provided workspace and style constraints.
3. Scenes: Define visual scenes to cover the whole timeline. Scene "id" starts at 0. Cover the whole timeline progressively. Do not overlap. IMPORTANT: Each scene MUST have a duration of EXACTLY 4, 6, or 8 seconds (e.g., 0-4, 4-10, 10-18) to match video generation API constraints. Choose the duration that best fits the pacing.
   - "type": Classify as "anchor" or "broll".
   - "intent": What the scene achieves (e.g., "problem_intro", "feature_demo", "transcription_visualization").
   - "visual": Give brief visual directions ("environment", "action"). Only one action per scene.
   - "camera": Must be an object with "shot" (e.g., "extreme_close_up", "medium_shot") and "movement" (e.g., "whip_pan", "static", "slow_push_in"). One movement per scene.
4. Beats: Subtitle Timing and Text. Lines must be progressive, non-overlapping, and continuously cover the ENTIRE duration with zero time gaps.
5. Reading Pace (Beats): Target 2.5-3.5 words/sec. Give long sentences adequate time and split them. Never cram text (e.g. 15 words in 0.7s).
6. Animations (Beats Subtitles): 'SCALE_IN' for the first beat, 'SLIDE_UP' for the rest.
7. Placement: Apply the requested placementStrategy. If DYNAMIC, choose TOP, CENTER, or BOTTOM per beat.
8. Emphasis: Pick 1-2 high-impact words per line (no filler words) existing in the text and add them to the 'emphasis' array.
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
					"duration": map[string]interface{}{"type": "number"},
					"visualDirection": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"workspace": map[string]interface{}{"type": "string"},
							"style":     map[string]interface{}{"type": "string"},
						},
						"required":             []string{"workspace", "style"},
						"additionalProperties": false,
					},
					"scenes": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"id":     map[string]interface{}{"type": "integer"},
								"type":   map[string]interface{}{"type": "string"},
								"intent": map[string]interface{}{"type": "string"},
								"start":  map[string]interface{}{"type": "number"},
								"end":    map[string]interface{}{"type": "number"},
								"visual": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"environment": map[string]interface{}{"type": "string"},
										"action":      map[string]interface{}{"type": "string"},
										"camera": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"shot":     map[string]interface{}{"type": "string"},
												"movement": map[string]interface{}{"type": "string"},
											},
											"required":             []string{"shot", "movement"},
											"additionalProperties": false,
										},
									},
									"required":             []string{"environment", "action", "camera"},
									"additionalProperties": false,
								},
							},
							"required":             []string{"id", "type", "intent", "start", "end", "visual"},
							"additionalProperties": false,
						},
					},
					"beats": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
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
							"required":             []string{"start", "end", "voice", "subtitle"},
							"additionalProperties": false,
						},
					},
				},
				"required":             []string{"duration", "visualDirection", "scenes", "beats"},
				"additionalProperties": false,
			},
		},
	}

	completionResp, err := u.client.Generate(ctx, systemPrompt, userPrompt, responseFormat, req.ProviderAPIKey)
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
