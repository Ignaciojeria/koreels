package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/reelgen/application/ports/out"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewGenerateKeyFramesUseCase)

type generateKeyFramesUseCase struct {
	chatClient out.ChatCompletionClient
}

func NewGenerateKeyFramesUseCase(chatClient out.ChatCompletionClient) in.GenerateKeyFramesExecutor {
	return &generateKeyFramesUseCase{chatClient: chatClient}
}

func (u *generateKeyFramesUseCase) Execute(ctx context.Context, req in.GenerateKeyFramesRequest) (in.GenerateKeyFramesResponse, error) {
	if len(req.Beats) == 0 {
		return in.GenerateKeyFramesResponse{}, fmt.Errorf("beats is required")
	}
	if req.ProviderAPIKey == "" {
		return in.GenerateKeyFramesResponse{}, fmt.Errorf("x-api-key header is required")
	}

	visualDir := req.VisualDirection
	if visualDir == nil {
		visualDir = u.generateVisualDirection(ctx, req.Beats, req.ProviderAPIKey)
	}
	if visualDir != nil && visualDir.AspectRatio == "" {
		visualDir.AspectRatio = "9:16"
	}

	beats := u.fillBeatPrompts(ctx, req.Beats, visualDir, req.ProviderAPIKey)

	return in.GenerateKeyFramesResponse{
		VoiceConfig:     req.VoiceConfig,
		Audio:           req.Audio,
		Beats:           beats,
		VisualDirection: visualDir,
	}, nil
}

// fillBeatPrompts genera un prompt visual por beat en UNA sola llamada al LLM.
func (u *generateKeyFramesUseCase) fillBeatPrompts(ctx context.Context, beats []in.Beat, visualDir *in.VisualDirection, apiKey string) []in.Beat {
	var narrations []string
	var indices []int
	for i, b := range beats {
		if b.Voice.Text != "" {
			narrations = append(narrations, b.Voice.Text)
			indices = append(indices, i)
		}
	}

	if len(narrations) == 0 {
		return beats
	}

	visualContext := buildVisualContext(visualDir)

	var beatList strings.Builder
	for i, n := range narrations {
		fmt.Fprintf(&beatList, "%d. \"%s\"\n", i+1, n)
	}

	fullScript := strings.Join(narrations, " ")

	systemPrompt := `You are an expert visual storyteller and prompt engineer for AI video generation (Vidu, Runway, Sora).
You will receive the FULL SCRIPT of a short-form social media reel, followed by individual beat narrations.
Your job: generate one video prompt per beat that together form a coherent visual narrative.

Narrative structure rules:
- Beat 1 must be a HOOK: visually striking, attention-grabbing, sets the tone
- Middle beats DEVELOP the story: show process, detail, or progression that matches the narration
- Final beat is the PAYOFF: resolution, satisfaction, call to action feeling
- Each beat's visual must directly illustrate what the narration says
- Maintain visual continuity: same characters/objects/environment should recur across beats
- Use progressive motion: camera angles and movement should evolve (e.g. wide→close→wide)

Technical rules:
- 40-60 words per prompt
- Include: subject, action, composition, lighting, color palette, mood, camera motion
- Optimize for vertical 9:16 format (portrait, social media reel)
- CRITICAL: Do NOT include any text, words, letters, or subtitles in the video. Purely visual.
- Output ONLY a JSON array of objects with "prompt" key
- Array length must match the number of beats
- No markdown, no explanation, just the JSON array`

	userPrompt := visualContext + fmt.Sprintf("FULL SCRIPT:\n%s\n\nBeats (%d total):\n%s", fullScript, len(narrations), beatList.String())

	resp, err := u.chatClient.Generate(ctx, systemPrompt, userPrompt, nil, apiKey)
	if err != nil || len(resp.Choices) == 0 {
		return beats
	}

	prompts := parseBeatPrompts(resp.Choices[0].Message.Content)

	result := make([]in.Beat, len(beats))
	copy(result, beats)
	for i, p := range prompts {
		if i >= len(indices) {
			break
		}
		result[indices[i]].Prompt = p.Prompt
	}

	return result
}

type beatPrompt struct {
	Prompt string `json:"prompt"`
}

func parseBeatPrompts(content string) []beatPrompt {
	content = strings.TrimSpace(content)
	content = stripMarkdownCodeBlock(content)

	var prompts []beatPrompt
	if err := json.Unmarshal([]byte(content), &prompts); err != nil {
		return nil
	}

	for i := range prompts {
		prompts[i].Prompt = sanitizePromptText(prompts[i].Prompt)
	}
	return prompts
}

func (u *generateKeyFramesUseCase) generateVisualDirection(ctx context.Context, beats []in.Beat, apiKey string) *in.VisualDirection {
	var narrations []string
	for _, b := range beats {
		if b.Voice.Text != "" {
			narrations = append(narrations, b.Voice.Text)
		}
	}
	if len(narrations) == 0 {
		return nil
	}

	script := strings.Join(narrations, " ")
	systemPrompt := `You are a creative director for short-form social video (reels).
Given the script, output a JSON object with exactly these keys:
- style (string): visual style, e.g. "cinematic tech noir"
- environment (string): setting, e.g. "modern dark workspace"
- lighting (string): lighting description, e.g. "dramatic rim lighting, high contrast"
- color_palette (array of strings): 3-5 colors, e.g. ["deep red", "dark gray", "neon blue"]
- camera_style (string): camera approach, e.g. "dynamic social media reel"
- aspect_ratio (string): must be "9:16"
Use concise values suitable for AI image generation prompts.
Output only valid JSON, no markdown or explanation.`

	resp, err := u.chatClient.Generate(ctx, systemPrompt, "Script:\n"+script, nil, apiKey)
	if err != nil || len(resp.Choices) == 0 {
		return nil
	}
	return parseVisualDirection(resp.Choices[0].Message.Content)
}

func buildVisualContext(dir *in.VisualDirection) string {
	if dir == nil {
		return ""
	}
	var parts []string
	if dir.Style != "" {
		parts = append(parts, "style: "+dir.Style)
	}
	if dir.Environment != "" {
		parts = append(parts, "environment: "+dir.Environment)
	}
	if dir.Lighting != "" {
		parts = append(parts, "lighting: "+dir.Lighting)
	}
	if dir.CameraStyle != "" {
		parts = append(parts, "camera: "+dir.CameraStyle)
	}
	if len(dir.ColorPalette) > 0 {
		parts = append(parts, "colors: "+strings.Join(dir.ColorPalette, ", "))
	}
	if len(parts) == 0 {
		return ""
	}
	return "Visual direction: " + strings.Join(parts, "; ") + ". "
}

func parseVisualDirection(content string) *in.VisualDirection {
	content = strings.TrimSpace(content)
	content = stripMarkdownCodeBlock(content)
	var dir in.VisualDirection
	if err := json.Unmarshal([]byte(content), &dir); err != nil {
		return nil
	}
	return &dir
}

func stripMarkdownCodeBlock(s string) string {
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	return s
}

func sanitizePromptText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"'`)
	var b strings.Builder
	lastSpace := false
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' || r == ' ' {
			if !lastSpace {
				b.WriteRune(' ')
				lastSpace = true
			}
			continue
		}
		lastSpace = false
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}
