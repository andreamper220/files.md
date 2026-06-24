package txt

import "strings"

// VoicePlaceholder is shown when a voice message has no transcript.
const VoicePlaceholder = "🎙 Голосовое сообщение"

// DisplayText returns human-readable text for previews, stripping media and
// preferring voice transcripts over the placeholder label.
func DisplayText(raw string) string {
	raw = strings.TrimSpace(NormNewLines(raw))
	if raw == "" {
		return ""
	}

	var fallback string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || HasImage(line) {
			continue
		}
		if line == VoicePlaceholder {
			if fallback == "" {
				fallback = line
			}
			continue
		}
		return line
	}

	if fallback != "" {
		return fallback
	}
	return strings.TrimSpace(strings.Split(raw, "\n")[0])
}
