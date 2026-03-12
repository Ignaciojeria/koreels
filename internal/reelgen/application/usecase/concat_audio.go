package usecase

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"time"

	"koreels/internal/reelgen/application/ports/in"
	"koreels/internal/reelgen/application/ports/out"

	"github.com/Ignaciojeria/ioc"
)

const wavHeaderSize = 44
const sampleRate = 24000
const numChannels = 1
const bitsPerSample = 16
const bytesPerSecond = sampleRate * numChannels * (bitsPerSample / 8)

var _ = ioc.Register(NewConcatAudioUseCase)

type concatAudioUseCase struct {
	uploader out.AudioUploader
	client   *http.Client
}

func NewConcatAudioUseCase(uploader out.AudioUploader) in.ConcatAudioExecutor {
	return &concatAudioUseCase{
		uploader: uploader,
		client:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (u *concatAudioUseCase) Execute(ctx context.Context, req in.ConcatAudioRequest) (in.ConcatAudioResponse, error) {
	var allPCM []byte
	var totalDuration float64

	for _, beat := range req.Beats {
		if beat.Voice.Audio == nil || beat.Voice.Audio.URL == "" {
			continue
		}
		wavBytes, err := u.downloadWAV(ctx, beat.Voice.Audio.URL)
		if err != nil {
			return in.ConcatAudioResponse{}, fmt.Errorf("beat %d: download: %w", beat.ID, err)
		}
		pcm, err := wavToPCM(wavBytes)
		if err != nil {
			return in.ConcatAudioResponse{}, fmt.Errorf("beat %d: parse wav: %w", beat.ID, err)
		}
		allPCM = append(allPCM, pcm...)
		totalDuration += beat.Voice.Audio.Duration
	}

	if len(allPCM) == 0 {
		return in.ConcatAudioResponse{}, fmt.Errorf("no audio URLs to concat")
	}

	wavBytes := pcmToWAV(allPCM)
	url, err := u.uploader.UploadWAV(ctx, wavBytes)
	if err != nil {
		return in.ConcatAudioResponse{}, fmt.Errorf("upload concat wav: %w", err)
	}

	// Si no teníamos duraciones en los beats, calcular desde PCM
	if totalDuration == 0 {
		totalDuration = float64(len(allPCM)) / float64(bytesPerSecond)
	}

	// Mismos beats que el request pero voice.audio sin URL (solo quitamos URLs individuales; duración se mantiene)
	beatsOut := make([]in.Beat, len(req.Beats))
	for i := range req.Beats {
		beatsOut[i] = req.Beats[i]
		if beatsOut[i].Voice.Audio != nil {
			beatsOut[i].Voice.Audio = &in.Audio{URL: "", Duration: beatsOut[i].Voice.Audio.Duration}
		}
	}

	return in.ConcatAudioResponse{
		Audio: in.ConcatAudioOutput{
			Voice: in.AudioTrack{
				URL:      url,
				Duration: totalDuration,
			},
		},
		Beats: beatsOut,
	}, nil
}

func (u *concatAudioUseCase) downloadWAV(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// wavToPCM extrae los bytes PCM de un WAV (asume header estándar 44 bytes).
func wavToPCM(wav []byte) ([]byte, error) {
	if len(wav) <= wavHeaderSize {
		return nil, fmt.Errorf("wav too short")
	}
	return wav[wavHeaderSize:], nil
}

func pcmToWAV(pcm []byte) []byte {
	dataSize := len(pcm)
	byteRate := sampleRate * numChannels * (bitsPerSample / 8)
	blockAlign := numChannels * (bitsPerSample / 8)
	fileSize := 4 + 24 + 8 + dataSize
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, []byte("RIFF"))
	binary.Write(buf, binary.LittleEndian, uint32(fileSize))
	binary.Write(buf, binary.LittleEndian, []byte("WAVE"))
	binary.Write(buf, binary.LittleEndian, []byte("fmt "))
	binary.Write(buf, binary.LittleEndian, uint32(16))
	binary.Write(buf, binary.LittleEndian, uint16(1))
	binary.Write(buf, binary.LittleEndian, uint16(numChannels))
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(buf, binary.LittleEndian, uint32(byteRate))
	binary.Write(buf, binary.LittleEndian, uint16(blockAlign))
	binary.Write(buf, binary.LittleEndian, uint16(bitsPerSample))
	binary.Write(buf, binary.LittleEndian, []byte("data"))
	binary.Write(buf, binary.LittleEndian, uint32(dataSize))
	buf.Write(pcm)
	return buf.Bytes()
}
