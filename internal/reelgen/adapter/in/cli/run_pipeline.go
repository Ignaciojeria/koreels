package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/shared/configuration"
	"koreels/internal/shared/infrastructure/observability"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.RegisterAtEnd(RunPipeline)

func RunPipeline(
	pipeline in.RunPipelineExecutor,
	conf configuration.Conf,
	obs observability.Observability,
) error {
	if conf.CLI_INPUT == "" {
		return fmt.Errorf("CLI_INPUT is required (set via --input flag)")
	}

	data, err := os.ReadFile(conf.CLI_INPUT)
	if err != nil {
		return fmt.Errorf("read input file %q: %w", conf.CLI_INPUT, err)
	}

	var req in.RunPipelineRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("parse input JSON: %w", err)
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("CLI_API_KEY")
	}
	req.ProviderAPIKey = apiKey

	obs.Logger.Info("cli_pipeline_start",
		"input", conf.CLI_INPUT,
		"output_dir", conf.OUTPUT_DIR,
		"script_len", len(req.ScriptText),
	)

	resp, err := pipeline.Execute(context.Background(), req)
	if err != nil {
		return fmt.Errorf("pipeline: %w", err)
	}

	output, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}

	fmt.Println(string(output))

	outputDir := conf.OUTPUT_DIR
	if outputDir == "" {
		outputDir = "."
	}
	jsonPath := filepath.Join(outputDir, "pipeline-output.json")
	if err := os.WriteFile(jsonPath, output, 0o644); err != nil {
		return fmt.Errorf("write output JSON to %s: %w", jsonPath, err)
	}

	obs.Logger.Info("cli_pipeline_complete",
		"audio_url", resp.Audio.Voice.URL,
		"audio_duration", resp.Audio.Voice.Duration,
		"beats", len(resp.Beats),
		"json_output", jsonPath,
	)

	return nil
}
