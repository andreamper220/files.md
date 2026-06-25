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
	normalized := NormalizeVoiceInNote(raw, false)
	if summary := VoiceSummary(normalized); summary != "" && isVoiceDominant(raw) {
		return summary
	}
	first := strings.TrimSpace(strings.Split(NormNewLines(normalized), "\n")[0])
	if first == VoicePlaceholder {
		return ""
	}
	return first
}

// FormatNoteDetailBody formats note content for detail views, expanding voice blocks.
func FormatNoteDetailBody(raw string) string {
	var body string
	if strings.Contains(raw, VoicePlaceholder) {
		body = NormalizeVoiceInNote(raw, true)
	} else if summary := VoiceSummary(raw); summary != "" {
		body = VoiceDetailBody(raw)
	} else {
		body = raw
	}
	return formatAttachmentLines(body)
}

func formatAttachmentLines(raw string) string {
	lines := strings.Split(NormNewLines(raw), "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if att, ok := ParseAttachmentLine(trimmed); ok {
			name := AttachmentDisplayName(att.Name, att.Path)
			out = append(out, "📎 ["+name+"]("+att.Path+")")
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// NormalizeVoiceInNote replaces voice placeholder blocks with summaries or detail bodies.
func NormalizeVoiceInNote(raw string, detail bool) string {
	lines := strings.Split(NormNewLines(raw), "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == VoicePlaceholder {
			voiceLines, mediaLine, endIdx := collectVoiceBlock(lines, i)
			mini := voiceBlockRaw(voiceLines, mediaLine)
			if detail {
				out = append(out, VoiceDetailBody(mini))
			} else if summary := VoiceSummary(mini); summary != "" {
				out = append(out, summary)
			}
			i = endIdx
			continue
		}
		if trimmed != "" && HasImage(trimmed) {
			if detail {
				out = append(out, trimmed)
			}
			continue
		}
		if att, ok := ParseAttachmentLine(trimmed); ok {
			name := AttachmentDisplayName(att.Name, att.Path)
			out = append(out, "📎 ["+name+"]("+att.Path+")")
			continue
		}
		out = append(out, lines[i])
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func collectVoiceBlock(lines []string, placeholderIdx int) (voiceLines []string, mediaLine string, endIdx int) {
	endIdx = placeholderIdx
	for j := placeholderIdx + 1; j < len(lines); j++ {
		t := strings.TrimSpace(lines[j])
		if t == "" {
			continue
		}
		if HasImage(t) {
			mediaLine = t
			endIdx = j
			return voiceLines, mediaLine, endIdx
		}
		voiceLines = append(voiceLines, t)
		endIdx = j
	}
	return voiceLines, mediaLine, endIdx
}

func voiceBlockRaw(voiceLines []string, mediaLine string) string {
	var parts []string
	if len(voiceLines) == 0 {
		parts = append(parts, VoicePlaceholder)
	} else {
		parts = append(parts, voiceLines...)
	}
	if mediaLine != "" {
		parts = append(parts, mediaLine)
	}
	return strings.Join(parts, "\n")
}

func isVoiceDominant(raw string) bool {
	for _, line := range strings.Split(NormNewLines(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || HasImage(line) || IsAttachmentLine(line) || line == VoicePlaceholder {
			continue
		}
		if summary := VoiceSummary(raw); summary != "" && line == summary {
			continue
		}
		return false
	}
	return VoiceSummary(raw) != "" || strings.Contains(raw, VoicePlaceholder)
}

// VoiceSummary returns the short title line for a voice note.
func VoiceSummary(raw string) string {
	summary, transcript := voiceTextLines(raw)
	if summary != "" && summary != VoicePlaceholder {
		return summary
	}
	if transcript != "" {
		first := strings.TrimSpace(strings.Split(transcript, "\n")[0])
		return SummarizeVoice(first)
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
		rest := textLines[1:]
		if len(rest) == 0 {
			return "", ""
		}
		if len(rest) == 1 {
			t := rest[0]
			return SummarizeVoice(t), t
		}
		return rest[0], strings.Join(rest[1:], "\n")
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
