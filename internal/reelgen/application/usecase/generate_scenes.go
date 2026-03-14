package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/reelgen/application/ports/out"

	"github.com/Ignaciojeria/ioc"
)

const MinSceneDurationSeconds = 4.0

var _ = ioc.Register(NewGenerateScenesUseCase)

type generateScenesUseCase struct {
	chatClient out.ChatCompletionClient
}

// NewGenerateScenesUseCase crea el caso de uso que agrupa beats en escenas ≥ 4s y opcionalmente genera prompts con LLM.
func NewGenerateScenesUseCase(chatClient out.ChatCompletionClient) in.GenerateScenesExecutor {
	return &generateScenesUseCase{
		chatClient: chatClient,
	}
}

func (u *generateScenesUseCase) Execute(ctx context.Context, req in.GenerateScenesRequest) (in.GenerateScenesResponse, error) {
	if err := u.validateBeats(req.Beats); err != nil {
		return in.GenerateScenesResponse{}, err
	}

	visualDir := req.VisualDirection
	if visualDir == nil && u.chatClient != nil && req.ProviderAPIKey != "" {
		visualDir = u.generateVisualDirection(ctx, req.Beats, req.ProviderAPIKey)
	}

	groups := u.groupBeatsIntoScenes(req.Beats)
	scenes := make([]in.Scene, 0, len(groups))

	for i, beatIDs := range groups {
		videoPrompt := ""
		if u.chatClient != nil && req.ProviderAPIKey != "" {
			videoPrompt = u.generateVideoPrompt(ctx, req.Beats, beatIDs, visualDir, req.ProviderAPIKey)
		}
		s := in.Scene{
			BeatIDs:     beatIDs,
			VideoPrompt: videoPrompt,
			Camera:      defaultCameraForScene(i, len(groups)),
		}
		if i < len(groups)-1 {
			s.TransitionOut = &in.TransitionOut{Type: "GLITCH", Duration: 0.2}
		}
		scenes = append(scenes, s)
	}

	return in.GenerateScenesResponse{
		Audio:           req.Audio,
		VisualDirection: visualDir,
		Beats:           req.Beats,
		Scenes:          scenes,
	}, nil
}

// validateBeats comprueba que no haya IDs duplicados y que tengan orden secuencial (por índice).
func (u *generateScenesUseCase) validateBeats(beats []in.Beat) error {
	seen := make(map[int]bool)
	for i, b := range beats {
		if seen[b.ID] {
			return fmt.Errorf("beat id duplicado: %d", b.ID)
		}
		seen[b.ID] = true
		if i > 0 && beats[i-1].ID >= b.ID {
			return fmt.Errorf("beats deben estar en orden secuencial por id; encontrado id %d después de %d", b.ID, beats[i-1].ID)
		}
	}
	return nil
}

// groupBeatsIntoScenes agrupa beats consecutivos de modo que cada escena sume ≥ MinSceneDurationSeconds.
// Si la última escena quedaría < 4s, se fusiona con la anterior.
func (u *generateScenesUseCase) groupBeatsIntoScenes(beats []in.Beat) [][]int {
	if len(beats) == 0 {
		return nil
	}

	var groups [][]int
	var current []int
	var duration float64

	for _, b := range beats {
		d := float64(0)
		if b.Voice.Audio != nil {
			d = b.Voice.Audio.Duration
		}
		current = append(current, b.ID)
		duration += d
		if duration >= MinSceneDurationSeconds {
			groups = append(groups, current)
			current = nil
			duration = 0
		}
	}

	if len(current) > 0 {
		if len(groups) > 0 && duration < MinSceneDurationSeconds {
			groups[len(groups)-1] = append(groups[len(groups)-1], current...)
		} else {
			groups = append(groups, current)
		}
	}

	return groups
}

func (u *generateScenesUseCase) sceneDuration(beats []in.Beat, beatIDs []int) float64 {
	byID := make(map[int]in.Beat)
	for _, b := range beats {
		byID[b.ID] = b
	}
	var d float64
	for _, id := range beatIDs {
		if b, ok := byID[id]; ok && b.Voice.Audio != nil {
			d += b.Voice.Audio.Duration
		}
	}
	return d
}

// generateVisualDirection genera una dirección visual global a partir de todos los beats (un solo llamado LLM).
func (u *generateScenesUseCase) generateVisualDirection(ctx context.Context, beats []in.Beat, apiKey string) *in.VisualDirection {
	var texts []string
	for _, b := range beats {
		if b.Voice.Text != "" {
			texts = append(texts, b.Voice.Text)
		}
	}
	if len(texts) == 0 {
		return nil
	}
	script := strings.Join(texts, " ")
	systemPrompt := `You are a creative director for short-form social video (reels). Given the script, output a JSON object with exactly these keys (all strings; color_palette is an array of strings): style, environment, lighting, color_palette, camera_style, aspect_ratio. Use concise values suitable for AI video generation. aspect_ratio must be "9:16". Output only valid JSON, no markdown or explanation.`
	userPrompt := "Script:\n" + script
	resp, err := u.chatClient.Generate(ctx, systemPrompt, userPrompt, nil, apiKey)
	if err != nil || len(resp.Choices) == 0 {
		return nil
	}
	// Parsear JSON a VisualDirection (simplificado: si falla, devolver un default)
	dir := parseVisualDirectionFromContent(resp.Choices[0].Message.Content)
	if dir != nil {
		if dir.AspectRatio == "" {
			dir.AspectRatio = "9:16"
		}
	}
	return dir
}

// generateVideoPrompt genera una sola frase de texto para el modelo de video (nunca JSON).
func (u *generateScenesUseCase) generateVideoPrompt(ctx context.Context, beats []in.Beat, beatIDs []int, visualDir *in.VisualDirection, apiKey string) string {
	byID := make(map[int]in.Beat)
	for _, b := range beats {
		byID[b.ID] = b
	}
	var texts []string
	for _, id := range beatIDs {
		if b, ok := byID[id]; ok {
			texts = append(texts, b.Voice.Text)
		}
	}
	if len(texts) == 0 {
		return ""
	}
	sceneText := strings.Join(texts, " ")
	globalContext := ""
	if visualDir != nil {
		parts := []string{}
		if visualDir.Style != "" {
			parts = append(parts, "style: "+visualDir.Style)
		}
		if visualDir.Environment != "" {
			parts = append(parts, "environment: "+visualDir.Environment)
		}
		if visualDir.Lighting != "" {
			parts = append(parts, "lighting: "+visualDir.Lighting)
		}
		if visualDir.CameraStyle != "" {
			parts = append(parts, "camera: "+visualDir.CameraStyle)
		}
		if len(parts) > 0 {
			globalContext = "Global visual: " + strings.Join(parts, ", ") + ". "
		}
	}
	systemPrompt := `You are a prompt writer for AI video generation. Output ONLY one short sentence in English (max 25 words) describing the visual for this clip. No JSON, no quotes, no markdown, no explanation. Just the sentence.`
	userPrompt := globalContext + "Scene narration: " + sceneText
	resp, err := u.chatClient.Generate(ctx, systemPrompt, userPrompt, nil, apiKey)
	if err != nil || len(resp.Choices) == 0 {
		return ""
	}
	return extractVideoPromptText(resp.Choices[0].Message.Content)
}

// extractVideoPromptText obtiene una frase usable para el modelo de video.
// Si el LLM devolvió JSON, intenta extraer "prompt", "video_prompt" o "description"; si no, sanitiza el texto.
func extractVideoPromptText(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// Quitar markdown code block si viene envuelto
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}

	// Si parece JSON, intentar extraer un campo de texto
	if strings.HasPrefix(s, "{") {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(s), &m); err == nil {
			for _, key := range []string{"prompt", "video_prompt", "description", "sentence", "text"} {
				if v, ok := m[key]; ok {
					if str, ok := v.(string); ok && strings.TrimSpace(str) != "" {
						return sanitizeVideoPrompt(str)
					}
				}
			}
			// Cualquier valor string en el primer nivel
			for _, v := range m {
				if str, ok := v.(string); ok && len(str) > 10 && !strings.HasPrefix(str, "{") {
					return sanitizeVideoPrompt(str)
				}
			}
		}
		return ""
	}
	return sanitizeVideoPrompt(s)
}

// sanitizeVideoPrompt deja solo texto apto para video (quita comillas, newlines).
func sanitizeVideoPrompt(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"'`)
	var b strings.Builder
	lastSpace := false
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			if !lastSpace {
				b.WriteRune(' ')
				lastSpace = true
			}
			continue
		}
		if unicode.IsSpace(r) {
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

// parseVisualDirectionFromContent intenta extraer un objeto VisualDirection del contenido (puede ser JSON o texto).
func parseVisualDirectionFromContent(content string) *in.VisualDirection {
	content = strings.TrimSpace(content)
	// Quitar markdown code block si viene envuelto
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}
	var dir in.VisualDirection
	if err := json.Unmarshal([]byte(content), &dir); err != nil {
		return nil
	}
	return &dir
}

// defaultCameraForScene devuelve una cámara por defecto según índice (variación suave entre escenas).
func defaultCameraForScene(index, total int) *in.CameraConfig {
	shots := []string{"medium close-up", "screen close-up", "macro interface view"}
	movements := []string{"slow push-in", "subtle parallax", "smooth pan"}
	i := index % len(shots)
	j := index % len(movements)
	return &in.CameraConfig{Shot: shots[i], Movement: movements[j]}
}
