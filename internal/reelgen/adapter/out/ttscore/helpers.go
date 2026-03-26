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

type PromptContext struct {
	FullScript string
	BeatIndex  int
	TotalBeats int
	PrevText   string
}

func BuildTTSPrompt(lang, style, text string, ctx *PromptContext) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	base := LanguageInstruction(lang)

	if ctx != nil && ctx.FullScript != "" && ctx.TotalBeats > 1 {
		return buildContextualPrompt(base, style, text, ctx)
	}

	if style != "" {
		style = strings.TrimSpace(style)
		style = strings.TrimSuffix(style, ":")
		return base + " " + style + ". " + text
	}
	return base + " " + text
}

func buildContextualPrompt(base, style, text string, ctx *PromptContext) string {
	var b strings.Builder

	b.WriteString(base)
	b.WriteString("\n")

	if style != "" {
		fmt.Fprintf(&b, "Delivery style: %s.\n", strings.TrimSuffix(strings.TrimSpace(style), ":"))
	}

	position := narrativePosition(ctx.BeatIndex, ctx.TotalBeats)
	fmt.Fprintf(&b, "You are narrating a short social media reel. This is line %d of %d (%s).\n",
		ctx.BeatIndex+1, ctx.TotalBeats, position)

	fmt.Fprintf(&b, "Full script for context (DO NOT read this aloud, only use it to understand tone and pacing):\n\"%s\"\n", ctx.FullScript)

	if ctx.PrevText != "" {
		fmt.Fprintf(&b, "The previous line was: \"%s\". Continue naturally from that energy and tone.\n", ctx.PrevText)
	}

	if ctx.BeatIndex == 0 {
		b.WriteString("Start with a confident, attention-grabbing tone. Do NOT begin with a pause.\n")
	} else {
		b.WriteString("Do NOT reset your energy. Continue as if this is the next breath in the same speech.\n")
	}

	fmt.Fprintf(&b, "Now speak ONLY this line:\n%s", text)
	return b.String()
}

func narrativePosition(index, total int) string {
	if total <= 1 {
		return "solo line"
	}
	if index == 0 {
		return "opening hook"
	}
	if index == total-1 {
		return "closing / call to action"
	}
	mid := total / 2
	if index < mid {
		return "building up"
	}
	return "delivering the payoff"
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
