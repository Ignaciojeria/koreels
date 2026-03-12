package in

import "context"

type GenerateBeatsRequest struct {
	ScriptText     string         `json:"scriptText"`
	LanguageCode   string         `json:"languageCode"`
	Subtitle       SubtitleConfig `json:"subtitle"`
	ProviderAPIKey string         `json:"-"` // Sent via header, not JSON body
}

type SubtitleConfig struct {
	Style             string `json:"style"`
	PlacementStrategy string `json:"placementStrategy"`
	MaxCharsPerLine   int    `json:"maxCharsPerLine"`
	MaxLines          int    `json:"maxLines"`
}

type GenerateBeatsResponse struct {
	Beats []Beat `json:"beats"`
}

type Beat struct {
	ID       int      `json:"id"`
	Voice    Voice    `json:"voice"`
	Subtitle Subtitle `json:"subtitle"`
}

type Voice struct {
	Text  string `json:"text"`
	Audio *Audio `json:"audio,omitempty"`
}

type Audio struct {
	URL      string  `json:"url"`
	Duration float64 `json:"duration"`
}

type Subtitle struct {
	Placement string `json:"placement"`
	Animation string `json:"animation"`
	Lines     []Line `json:"lines"`
}

type Line struct {
	Text     string   `json:"text"`
	Emphasis []string `json:"emphasis,omitempty"`
}

type GenerateBeatsExecutor interface {
	Execute(ctx context.Context, req GenerateBeatsRequest) (GenerateBeatsResponse, error)
}
