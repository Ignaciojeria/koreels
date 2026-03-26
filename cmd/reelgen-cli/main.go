// reelgen-cli runs the reelgen pipeline locally without GCS dependencies.
// Only GEMINI_API_KEY is required. WAV files are saved to the output directory.
//
// Usage:
//
//	go run ./cmd/reelgen-cli --input pipeline.json --output ./output [--api-key sk-xxx]
package main

import (
	"flag"
	"log"
	"os"
	"strings"

	reelgen "koreels"

	_ "koreels/internal/reelgen/adapter/in/cli"
	_ "koreels/internal/reelgen/adapter/out/chatcompletion"
	_ "koreels/internal/reelgen/adapter/out/geminiapi"
	_ "koreels/internal/reelgen/adapter/out/local_uploader"
	_ "koreels/internal/reelgen/adapter/out/localtts"
	_ "koreels/internal/reelgen/adapter/out/novideo"
	_ "koreels/internal/reelgen/adapter/out/qwenapi"
	_ "koreels/internal/reelgen/application/usecase"
	_ "koreels/internal/shared/configuration"
	_ "koreels/internal/shared/infrastructure/ai"
	_ "koreels/internal/shared/infrastructure/observability"

	"github.com/Ignaciojeria/ioc"
)

func main() {
	inputFile := flag.String("input", "", "Path to the pipeline input JSON file (required)")
	outputDir := flag.String("output", "./output", "Directory to save WAV files")
	apiKey := flag.String("api-key", "", "Gemini API key (overrides GEMINI_API_KEY env var)")
	flag.Parse()

	if *inputFile == "" {
		log.Fatal("--input flag is required")
	}

	os.Setenv("VERSION", strings.TrimSpace(reelgen.Version))
	os.Setenv("CLI_INPUT", *inputFile)
	os.Setenv("OUTPUT_DIR", *outputDir)

	if *apiKey != "" {
		os.Setenv("GEMINI_API_KEY", *apiKey)
	}

	if err := ioc.LoadDependencies(); err != nil {
		log.Fatal("Failed to load dependencies: ", err)
	}
}
