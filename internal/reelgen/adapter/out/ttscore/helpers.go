package ttscore

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"

	"google.golang.org/genai"
)


const TTSModel = "gemini-2.5-flash-preview-tts"
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
	langName := languageName(lang)

	var b strings.Builder

	b.WriteString("# AUDIO PROFILE: Narrator\n")
	b.WriteString("## \"Social Media Reel Narrator\"\n\n")

	b.WriteString("## THE SCENE: Recording Studio\n")
	fmt.Fprintf(&b, "A professional podcast studio with a high-end condenser microphone. The narrator is recording a short-form social media reel in %s. The energy is focused and engaging.\n\n", langName)

	b.WriteString("### DIRECTOR'S NOTES\n")

	if style != "" {
		fmt.Fprintf(&b, "Style: %s.\n", strings.TrimSuffix(strings.TrimSpace(style), ":"))
	} else {
		b.WriteString("Style: Confident, engaging, and dynamic. Use a vocal smile.\n")
	}

	b.WriteString("Pacing: Speak at a fast, consistent pace throughout. No pauses at the start. Keep the tempo steady and energetic like a professional short-form content creator.\n")
	fmt.Fprintf(&b, "Accent: Natural %s accent.\n\n", langName)

	b.WriteString("#### TRANSCRIPT\n")
	b.WriteString(text)

	return b.String()
}

func languageName(lang string) string {
	if strings.HasPrefix(lang, "es") {
		if lang == "es-es" || strings.HasPrefix(lang, "es-es") {
			return "Spanish (Spain)"
		}
		if strings.Contains(lang, "es-") {
			return "Spanish (Latin American)"
		}
		return "Spanish"
	}
	switch {
	case strings.HasPrefix(lang, "en"):
		return "English"
	case strings.HasPrefix(lang, "pt"):
		return "Portuguese (Brazilian)"
	case strings.HasPrefix(lang, "fr"):
		return "French"
	case strings.HasPrefix(lang, "de"):
		return "German"
	case strings.HasPrefix(lang, "it"):
		return "Italian"
	case strings.HasPrefix(lang, "ja"):
		return "Japanese"
	default:
		return "the appropriate language"
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
