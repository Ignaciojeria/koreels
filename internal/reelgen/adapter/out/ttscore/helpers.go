package ttscore

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"

	"google.golang.org/genai"
)


const TTSModel = "gemini-2.5-pro-preview-tts"
const DefaultVoice = "Kore"
const DefaultLanguage = "es-MX"
const BytesPerSecond = 24000 * 2 // PCM s16le 24kHz mono

func SynthesizeToPCM(ctx context.Context, client *genai.Client, promptText string, voiceName string) ([]byte, error) {
	contents := []*genai.Content{
		{Role: "user", Parts: []*genai.Part{{Text: promptText}}},
	}
	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{string(genai.ModalityAudio)},
		SpeechConfig: &genai.SpeechConfig{
			VoiceConfig: &genai.VoiceConfig{
				PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
					VoiceName: voiceName,
				},
			},
		},
	}
	resp, err := client.Models.GenerateContent(ctx, TTSModel, contents, config)
	if err != nil {
		return nil, fmt.Errorf("synthesizing speech: %w", err)
	}
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, fmt.Errorf("no audio generated (empty response)")
	}
	var pcmBytes []byte
	for _, p := range resp.Candidates[0].Content.Parts {
		if p != nil && p.InlineData != nil && len(p.InlineData.Data) > 0 {
			pcmBytes = p.InlineData.Data
			break
		}
	}
	if len(pcmBytes) == 0 {
		return nil, fmt.Errorf("synthesized audio is empty")
	}
	return pcmBytes, nil
}

func BuildTTSPrompt(lang, style, text string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	base := LanguageInstruction(lang)
	if style != "" {
		style = strings.TrimSpace(style)
		style = strings.TrimSuffix(style, ":")
		return base + " " + style + ". " + text
	}
	return base + " " + text
}

func LanguageInstruction(lang string) string {
	if strings.HasPrefix(lang, "es") {
		if lang == "es-es" || strings.HasPrefix(lang, "es-es") {
			return "Speak in Spanish (Spain). Use clear, natural intonation and modulation."
		}
		if lang == "es-mx" || lang == "es-ar" || lang == "es-co" || lang == "es-cl" || strings.Contains(lang, "es-") {
			return "Speak in Spanish (Latin American). Use clear, natural intonation and modulation."
		}
		return "Speak in Spanish with clear, natural intonation and modulation."
	}
	switch {
	case strings.HasPrefix(lang, "en"):
		return "Speak in English with clear, natural intonation and modulation."
	case strings.HasPrefix(lang, "pt"):
		return "Speak in Portuguese with clear, natural intonation and modulation."
	case strings.HasPrefix(lang, "fr"):
		return "Speak in French with clear, natural intonation and modulation."
	case strings.HasPrefix(lang, "de"):
		return "Speak in German with clear, natural intonation and modulation."
	case strings.HasPrefix(lang, "it"):
		return "Speak in Italian with clear, natural intonation and modulation."
	case strings.HasPrefix(lang, "ja"):
		return "Speak in Japanese with clear, natural intonation and modulation."
	default:
		return "Speak with clear, natural intonation and modulation."
	}
}

func PcmToWAV(pcm []byte, sampleRate, numChannels, bitsPerSample int) []byte {
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
