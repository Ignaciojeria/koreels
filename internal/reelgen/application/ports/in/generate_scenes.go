package in

import "context"

type GenerateScenesRequest struct {
	ScriptText        string          `json:"scriptText"`
	VisualDirection   VisualDirection `json:"visualDirection"`
	SubtitleStyle     string          `json:"subtitleStyle"`
	PlacementStrategy string          `json:"placementStrategy"`
	LanguageCode      string          `json:"languageCode"`
	MaxCharsPerLine   int             `json:"maxCharsPerLine"`
	MaxLines          int             `json:"maxLines"`
	OutputMediaType   string          `json:"outputMediaType"`
	ProviderAPIKey    string          `json:"-"` // Sent via header, not JSON body
}

type GenerateScenesResponse struct {
	Duration        float64         `json:"duration"`
	VisualDirection VisualDirection `json:"visualDirection"`
	Scenes          []Scene         `json:"scenes"`
	Beats           []Beat          `json:"beats"`
}

type VisualDirection struct {
	Workspace string `json:"workspace"`
	Style     string `json:"style"`
}

type Scene struct {
	ID     int     `json:"id"`
	Type   string  `json:"type"`
	Intent string  `json:"intent"`
	Start  float64 `json:"start"`
	End    float64 `json:"end"`
	Visual Visual  `json:"visual"`
}

type Visual struct {
	Environment string `json:"environment"`
	Action      string `json:"action"`
	Camera      Camera `json:"camera"`
}

type Camera struct {
	Shot     string `json:"shot"`
	Movement string `json:"movement"`
}

type Beat struct {
	Start    float64  `json:"start"`
	End      float64  `json:"end"`
	Voice    Voice    `json:"voice"`
	Subtitle Subtitle `json:"subtitle"`
}

type Voice struct {
	Text string `json:"text"`
}

type Subtitle struct {
	Placement string `json:"placement"`
	Animation string `json:"animation"`
	Lines     []Line `json:"lines"`
}

type Line struct {
	Text     string   `json:"text"`
	Start    float64  `json:"start"`
	End      float64  `json:"end"`
	Emphasis []string `json:"emphasis,omitempty"`
}

type GenerateScenesExecutor interface {
	Execute(ctx context.Context, req GenerateScenesRequest) (GenerateScenesResponse, error)
}
