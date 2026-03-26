package local_uploader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"koreels/internal/reelgen/application/ports/out"
	"koreels/internal/shared/configuration"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewLocalAudioUploader)

type localAudioUploader struct {
	outputDir string
}

func NewLocalAudioUploader(conf configuration.Conf) (out.AudioUploader, error) {
	dir := conf.OUTPUT_DIR
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create output dir %q: %w", dir, err)
	}
	return &localAudioUploader{outputDir: dir}, nil
}

func (a *localAudioUploader) UploadWAV(ctx context.Context, wavBytes []byte) (string, error) {
	fileName := fmt.Sprintf("concat_%d.wav", time.Now().UnixNano())
	filePath := filepath.Join(a.outputDir, fileName)
	if err := os.WriteFile(filePath, wavBytes, 0o644); err != nil {
		return "", fmt.Errorf("write concat wav to %s: %w", filePath, err)
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}
	return absPath, nil
}
