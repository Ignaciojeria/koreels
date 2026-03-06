package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Ignaciojeria/ioc"
	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/reelgen/application/ports/out"
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
	systemPrompt := `You are an expert video editing assistant. 
Your task is to take a given script and segment it into coherent scenes for a short video (like Instagram Reels, TikTok).
Follow these critical rules:
1. Dynamic Subtitles (Relative Time): Subtitles must use relative timestamps starting from 0.0 for each scene. Lines within a scene must NOT overlap in time. They must appear progressively (e.g., Line 1: 0.0-1.5, Line 2: 1.5-3.2). 
2. No Dead Time: The lines MUST cover the ENTIRE scene duration continuously. The first line's start MUST be 0.0. The last line's end MUST equal the scene's total duration (end - start). There must be NO time gaps between lines.
3. Strict Reading Pace & Line Length: Ensure a reading speed of 2.5 to 3.5 words per second (approx 14-18 characters per second). NEVER cram long sentences into short durations (e.g., do not put 15 words in 0.7s). If a sentence is long, allocate enough time (e.g., 4-5 seconds) and split it into multiple lines.
4. No Empty Lines: Never output empty strings in the 'lines' array. Only include real text.
5. Consistent Animations: Use 'SCALE_IN' for the hook (index 0) and 'SLIDE_UP' for all other scenes.
6. Emphasis: Identify 1 or 2 high-impact words per line to be visually emphasized in the video. Put them in the 'emphasis' array. Only include words that actually exist in the line text. Do not emphasize filler words.
Respect the user's constraints: subtitle style, placement, max chars per line, and max lines. You must strictly return a JSON object matching the required schema.`

	userPromptBytes, _ := json.Marshal(req)
	userPrompt := fmt.Sprintf("Please process the following request parameters and script:\n%s", string(userPromptBytes))

	responseFormat := map[string]interface{}{
		"type": "json_schema",
		"json_schema": map[string]interface{}{
			"name": "generate_scenes_response",
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
														"type": "array",
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
