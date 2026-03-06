package in

import "context"

type GenerateScenesRequest struct {
	ScriptText        string `json:"scriptText"`
	SubtitleStyle     string `json:"subtitleStyle"`
	PlacementStrategy string `json:"placementStrategy"`
	LanguageCode      string `json:"languageCode"`
	MaxCharsPerLine   int    `json:"maxCharsPerLine"`
	MaxLines          int    `json:"maxLines"`
	OutputMediaType   string `json:"outputMediaType"`
}

type GenerateScenesResponse struct {
	Scenes []Scene `json:"scenes"`
}

type Scene struct {
	Index    int      `json:"index"`
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
