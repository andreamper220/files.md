package priority

import (
	"strings"
	"unicode/utf8"
)

var DefaultEmojis = []string{"🔴", "🟠", "🟡", "🟢", "⚪️", "🔵"}

// Strip removes a leading configured priority emoji from task text.
func Strip(text string, emojis []string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return text
	}
	for _, emoji := range emojis {
		if strings.HasPrefix(text, emoji) {
			return strings.TrimSpace(strings.TrimPrefix(text, emoji))
		}
	}
	r, size := utf8.DecodeRuneInString(text)
	if r != utf8.RuneError && size > 0 && !isLetterOrDigit(r) {
		return strings.TrimSpace(text[size:])
	}
	return text
}

// Apply prefixes text with the selected priority emoji.
func Apply(text, emoji string, emojis []string) string {
	text = Strip(text, emojis)
	if emoji == "" {
		return text
	}
	if text == "" {
		return emoji
	}
	return emoji + " " + text
}

// Detect returns the leading priority emoji if present.
func Detect(text string, emojis []string) string {
	text = strings.TrimSpace(text)
	for _, emoji := range emojis {
		if strings.HasPrefix(text, emoji) {
			return emoji
		}
	}
	return ""
}

func isLetterOrDigit(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || (r >= 'А' && r <= 'я') || r == 'Ё' || r == 'ё'
}
