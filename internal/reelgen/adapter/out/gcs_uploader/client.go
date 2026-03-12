package gcs_uploader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"koreels/internal/reelgen/application/ports/out"
	"koreels/internal/shared/configuration"

	gcs "cloud.google.com/go/storage"
	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewAudioUploader)

type audioUploader struct {
	gcsClient *gcs.Client
	conf      configuration.Conf
}

// NewAudioUploader returns an AudioUploader that uploads WAV to GCS (reelgen/audio/).
func NewAudioUploader(gcsClient *gcs.Client, conf configuration.Conf) (out.AudioUploader, error) {
	return &audioUploader{gcsClient: gcsClient, conf: conf}, nil
}

func (a *audioUploader) UploadWAV(ctx context.Context, wavBytes []byte) (string, error) {
	if a.gcsClient == nil || a.conf.GCS_BUCKET == "" {
		return "", fmt.Errorf("GCS client or GCS_BUCKET not configured")
	}
	bucketName := a.conf.GCS_BUCKET
	fileName := fmt.Sprintf("%d.wav", time.Now().UnixNano())
	objectPath := fmt.Sprintf("reelgen/audio/%s", fileName)

	uploadURL, err := a.gcsClient.Bucket(bucketName).SignedURL(objectPath, &gcs.SignedURLOptions{
		Method:      "PUT",
		Expires:     time.Now().Add(15 * time.Minute),
		ContentType: "audio/wav",
	})
	if err != nil {
		return "", fmt.Errorf("signed url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(wavBytes))
	if err != nil {
		return "", fmt.Errorf("create upload request: %w", err)
	}
	req.Header.Set("Content-Type", "audio/wav")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("upload status %d (bucket=%q): %s", resp.StatusCode, bucketName, string(body))
	}

	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectPath), nil
}
