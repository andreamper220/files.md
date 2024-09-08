package txt

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func I64(i int64) string {
	return strconv.FormatInt(i, 10)
}

func Ucfirst(str string) string {
	for _, v := range str {
		u := string(unicode.ToUpper(v))
		return u + str[len(u):]
	}
	return ""
}

func Lcfirst(str string) string {
	for _, v := range str {
		u := string(unicode.ToLower(v))
		return u + str[len(u):]
	}
	return ""
}

// Substr isn't multi-Unicode-codepoint aware, like specifying skintone or
// gender of an emoji: https://unicode.org/emoji/charts/full-emoji-modifiers.html
func Substr(input string, start int, length int) string {
	asRunes := []rune(input)
	if start >= len(asRunes) {
		return ""
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}

	return string(asRunes[start : start+length])
}

func Emoji(emoji, str string) string {
	if emoji == "" {
		return str
	}

	return fmt.Sprintf("%s %s", emoji, str)
}

func NormNewLines(text string) string {
	text = strings.Replace(text, "\\r\\n", "\n", -1)
	return strings.Replace(text, "\\n\\r", "\n", -1)
}

// Spaces-like characters are trimmed out
// TODO add tests
func SplitTextIntoChunks(text string, maxLen int) []string {
	var chunks []string

	for utf8.RuneCountInString(text) > maxLen {
		// Get the substring of the first maxLen runes
		runes := []rune(text)
		subStr := string(runes[:maxLen])

		// Find the last newline in the substring
		splitIndex := strings.LastIndex(subStr, "\n")
		if splitIndex == -1 {
			// No newline found, find the last space
			splitIndex = strings.LastIndex(subStr, " ")
			if splitIndex == -1 {
				// No space found either, split at maxLen
				splitIndex = maxLen
			}
		} else {
			// Adjust the split index to the rune count
			splitIndex = utf8.RuneCountInString(subStr[:splitIndex])
		}

		chunks = append(chunks, strings.TrimSpace(string(runes[:splitIndex])))
		text = string(runes[splitIndex:])
	}
	chunks = append(chunks, strings.TrimSpace(text))

	return chunks
}

func InsertTextAfterHeader(existingContent, header, newContent string) string {
	if !strings.Contains(existingContent, header) {
		return strings.TrimSpace(fmt.Sprintf("%s\n%s\n%s", header, newContent, existingContent))
	}

	headerAndContent := fmt.Sprintf("%s\n%s", header, newContent)
	content := strings.Replace(existingContent, header, headerAndContent, 1)

	return strings.TrimSpace(content)
}

// TODO add tests
func FirstWord(str string) string {
	str = strings.TrimSpace(str)
	re := regexp.MustCompile(`^[^\s\p{P}]+`)
	return re.FindString(str)
}

// TODO add tests
func EscapeHTML(str string) string {
	// HTML escaping
	var htmlEscaper = strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)

	return htmlEscaper.Replace(str)
}

// TODO add tests
func ReplaceWithPlaceholders(str, regex, placeholder string) (string, map[string]string) {
	re := regexp.MustCompile(regex)
	placeholders := make(map[string]string)
	counter := 0

	// Function to replace each match with a placeholder
	res := re.ReplaceAllStringFunc(str, func(match string) string {
		p := fmt.Sprintf("#%s%d#", placeholder, counter)
		placeholders[p] = match
		counter++
		return p
	})

	return res, placeholders
}

func RestoreFromPlaceholders(str string, placeholders map[string]string) string {
	for placeholder, original := range placeholders {
		str = strings.ReplaceAll(str, placeholder, original)
	}
	return str
}
