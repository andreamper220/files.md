package txt

import (
	"fmt"
	"path/filepath"
	"strings"
)

// closeParenForMarkdownURL finds the closing ")" for a markdown link URL.
func closeParenForMarkdownURL(rest string) int {
	first := strings.Index(rest, ")")
	if first == -1 {
		return -1
	}
	candidate := rest[:first]
	if strings.Contains(candidate, "(") {
		return strings.LastIndex(rest, ")")
	}
	return first
}
func parseURLInParens(s string) (url string, ok bool) {
	s = strings.TrimSpace(s)
	if s == "" || s[len(s)-1] != ')' {
		return "", false
	}
	return strings.TrimSpace(s[:len(s)-1]), true
}

// parseBracketLink parses [label](url), including ")" inside the URL.
func parseBracketLink(s string) (label, url string, ok bool) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[") {
		return "", "", false
	}
	bracket := strings.Index(s, "](")
	if bracket == -1 {
		return "", "", false
	}
	label = strings.TrimSpace(s[1:bracket])
	url, ok = parseURLInParens(s[bracket+2:])
	return label, url, ok
}

// ParseImageLine parses ![](path), including ")" in filenames.
func ParseImageLine(line string) (path string, ok bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "![") {
		return "", false
	}
	bracket := strings.Index(line, "](")
	if bracket == -1 {
		return "", false
	}
	path, ok = parseURLInParens(line[bracket+2:])
	return strings.TrimPrefix(strings.TrimSpace(path), "/"), ok
}

const attachmentPlaceholder = "·"

// AttachmentInfo describes a saved file attachment embedded in markdown.
type AttachmentInfo struct {
	Name string
	Path string
}

// FormatAttachmentContent returns markdown for a saved file with a display name.
func FormatAttachmentContent(mediaPath, name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Sprintf("📎 [%s](%s)", attachmentPlaceholder, mediaPath)
	}
	return fmt.Sprintf("📎 [%s](%s)", name, mediaPath)
}

// IsAttachmentLine reports whether a line is a file attachment link.
func IsAttachmentLine(line string) bool {
	_, ok := ParseAttachmentLine(line)
	return ok
}

// ParseAttachmentLine parses a single attachment markdown line.
func ParseAttachmentLine(line string) (AttachmentInfo, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "📎") {
		return AttachmentInfo{}, false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(line, "📎"))
	name, path, ok := parseBracketLink(rest)
	if !ok {
		return AttachmentInfo{}, false
	}
	return AttachmentInfo{Name: name, Path: path}, true
}

// ParseAttachment extracts the first attachment link from markdown content.
func ParseAttachment(raw string) (AttachmentInfo, bool) {
	for _, line := range strings.Split(NormNewLines(raw), "\n") {
		if att, ok := ParseAttachmentLine(line); ok {
			return att, true
		}
	}
	return AttachmentInfo{}, false
}

// ParseAttachments returns every attachment link in markdown content.
func ParseAttachments(raw string) []AttachmentInfo {
	var out []AttachmentInfo
	for _, line := range strings.Split(NormNewLines(raw), "\n") {
		if att, ok := ParseAttachmentLine(line); ok {
			out = append(out, att)
		}
	}
	return out
}

// AttachmentNames returns display names for all attachments in content.
func AttachmentNames(raw string) []string {
	attachments := ParseAttachments(raw)
	if len(attachments) == 0 {
		return nil
	}
	names := make([]string, 0, len(attachments))
	for _, att := range attachments {
		names = append(names, AttachmentDisplayName(att.Name, att.Path))
	}
	return names
}

// AttachmentNoteTitle builds a note title from one or more attachment names.
func AttachmentNoteTitle(names []string) string {
	if len(names) == 0 {
		return ""
	}
	if len(names) == 1 {
		return names[0]
	}
	return strings.Join(names, " + ")
}

// AttachmentMediaPath returns media/filename for a stored attachment path.
func AttachmentMediaPath(mediaPath string) (dir, filename string) {
	mediaPath = strings.TrimPrefix(strings.TrimSpace(mediaPath), "/")
	parts := strings.Split(mediaPath, "/")
	if len(parts) == 1 {
		return "media", parts[0]
	}
	return strings.Join(parts[:len(parts)-1], "/"), parts[len(parts)-1]
}

// AttachmentDisplayName returns a human-readable attachment name.
func AttachmentDisplayName(name, mediaPath string) string {
	name = strings.TrimSpace(name)
	if name != "" && name != attachmentPlaceholder {
		return name
	}
	return filepath.Base(mediaPath)
}

// NeedsUserTitle reports whether the user should name this draft (media-only).
func NeedsUserTitle(raw string) bool {
	if DraftTitle(raw) != "" {
		return false
	}
	return IsAttachmentOnly(raw)
}

// IsAttachmentOnly reports whether content contains only file attachments (optional title line).
func IsAttachmentOnly(raw string) bool {
	title := DraftTitle(raw)
	hasAttachment := false
	for _, line := range strings.Split(NormNewLines(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if _, ok := ParseAttachmentLine(line); ok {
			hasAttachment = true
			continue
		}
		if HasImage(line) || line == VoicePlaceholder {
			return false
		}
		if title != "" && line == title {
			continue
		}
		return false
	}
	return hasAttachment
}

// DraftTitle returns the first user-authored title line in a draft.
func DraftTitle(raw string) string {
	for _, line := range strings.Split(NormNewLines(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || HasImage(line) || IsAttachmentLine(line) {
			continue
		}
		if line == VoicePlaceholder {
			continue
		}
		return line
	}
	return ""
}

// ReplaceNote replaces entire note content with newContent.
func ReplaceNote(newContent string) string {
	newContent = strings.TrimSpace(newContent)
	if newContent == "" {
		return ""
	}
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	return newContent
}

// AppendNoteContent appends text or attachment markdown to a note.
func AppendNoteContent(raw, addition string) string {
	raw = strings.TrimRight(raw, "\n")
	addition = strings.TrimSpace(addition)
	if addition == "" {
		if raw == "" {
			return ""
		}
		return raw + "\n"
	}
	if raw == "" {
		return addition + "\n"
	}
	return raw + "\n" + addition + "\n"
}

// HasNoteMedia reports whether a note contains file or image attachments.
func HasNoteMedia(raw string) bool {
	for _, line := range strings.Split(NormNewLines(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if IsAttachmentLine(trimmed) || HasImage(trimmed) {
			return true
		}
	}
	return false
}

// ApplyDraftTitle prepends a user title and labels unnamed attachment links.
func ApplyDraftTitle(content, title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return content
	}
	var lines []string
	for _, line := range strings.Split(NormNewLines(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if att, ok := ParseAttachmentLine(trimmed); ok {
			label := strings.TrimSpace(att.Name)
			if label == "" || label == attachmentPlaceholder {
				line = FormatAttachmentContent(att.Path, title)
			}
		}
		lines = append(lines, line)
	}
	content = strings.Join(lines, "\n")
	if DraftTitle(content) == title {
		return content
	}
	return title + "\n" + strings.TrimSpace(content)
}
