package in

import "context"

type GenerateAudioRequest struct {
	Beats          []Beat `json:"beats"`
	ProviderAPIKey string `json:"-"` // Just in case it's needed via header
}

type GenerateAudioResponse struct {
	Beats []Beat `json:"beats"`
}

type GenerateAudioExecutor interface {
	Execute(ctx context.Context, req GenerateAudioRequest) (GenerateAudioResponse, error)
}
