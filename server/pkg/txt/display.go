package txt

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// VoicePlaceholder is shown when a voice message has no transcript.
const VoicePlaceholder = "🎙 Голосовое сообщение"

const voiceSummaryMaxRunes = 80

// DisplayText returns human-readable text for previews, stripping media and
// preferring voice summaries over the placeholder label.
func DisplayText(raw string) string {
	if summary := VoiceSummary(raw); summary != "" {
		return summary
	}
	first := strings.TrimSpace(strings.Split(NormNewLines(raw), "\n")[0])
	if first == VoicePlaceholder {
		return ""
	}
	return first
}

// VoiceSummary returns the short title line for a voice note.
func VoiceSummary(raw string) string {
	summary, _ := voiceTextLines(raw)
	if summary != "" && summary != VoicePlaceholder {
		return summary
	}
	return ""
}

// VoiceTranscript returns the full transcription for a voice note.
func VoiceTranscript(raw string) string {
	summary, transcript := voiceTextLines(raw)
	if transcript != "" {
		return transcript
	}
	if summary != "" && summary != VoicePlaceholder {
		return summary
	}
	return ""
}

// VoiceDetailBody returns summary + transcript for detail views.
func VoiceDetailBody(raw string) string {
	summary, transcript := voiceTextLines(raw)
	if summary == "" || summary == VoicePlaceholder {
		return raw
	}
	if transcript == "" || transcript == summary {
		return summary + voiceMediaSuffix(raw)
	}
	return summary + "\n\n" + transcript + voiceMediaSuffix(raw)
}

func voiceMediaSuffix(raw string) string {
	var media []string
	for _, line := range strings.Split(NormNewLines(raw), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && HasImage(line) {
			media = append(media, line)
		}
	}
	if len(media) == 0 {
		return ""
	}
	return "\n\n" + strings.Join(media, "\n")
}

func voiceTextLines(raw string) (summary, transcript string) {
	var textLines []string
	for _, line := range strings.Split(NormNewLines(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || HasImage(line) {
			continue
		}
		textLines = append(textLines, line)
	}
	if len(textLines) == 0 {
		return "", ""
	}
	if len(textLines) == 1 {
		return textLines[0], ""
	}
	if textLines[0] == VoicePlaceholder {
		return "", strings.Join(textLines[1:], "\n")
	}
	return textLines[0], textLines[1]
}

// SummarizeVoice shortens a transcript for use as a preview title.
func SummarizeVoice(transcript string) string {
	transcript = strings.TrimSpace(transcript)
	if transcript == "" {
		return ""
	}
	if utf8.RuneCountInString(transcript) <= voiceSummaryMaxRunes {
		return transcript
	}
	if idx := strings.IndexAny(transcript, ".!?"); idx > 0 {
		sentence := strings.TrimSpace(transcript[:idx+1])
		if utf8.RuneCountInString(sentence) <= voiceSummaryMaxRunes*2 {
			return sentence
		}
	}
	return Substr(transcript, 0, voiceSummaryMaxRunes-1) + "…"
}

// FormatVoiceContent stores summary + optional full transcript before media.
func FormatVoiceContent(transcript string) string {
	transcript = strings.TrimSpace(transcript)
	if transcript == "" {
		return VoicePlaceholder
	}
	transcript = Ucfirst(transcript)
	summary := SummarizeVoice(transcript)
	if summary == transcript {
		return transcript
	}
	return summary + "\n" + transcript
}

// IsEmojiOnlyName reports whether a display name contains only emoji and spaces.
func IsEmojiOnlyName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
