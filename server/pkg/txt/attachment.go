package txt

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var attachmentRE = regexp.MustCompile(`📎\s*\[([^\]]*)\]\(([^)]+)\)`)

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
	m := attachmentRE.FindStringSubmatch(strings.TrimSpace(line))
	if m == nil {
		return AttachmentInfo{}, false
	}
	return AttachmentInfo{
		Name: strings.TrimSpace(m[1]),
		Path: strings.TrimSpace(m[2]),
	}, true
}

// ParseAttachment extracts the first attachment link from markdown content.
func ParseAttachment(raw string) (AttachmentInfo, bool) {
	m := attachmentRE.FindStringSubmatch(raw)
	if m == nil {
		return AttachmentInfo{}, false
	}
	return AttachmentInfo{
		Name: strings.TrimSpace(m[1]),
		Path: strings.TrimSpace(m[2]),
	}, true
}

// ParseAttachments returns every attachment link in markdown content.
func ParseAttachments(raw string) []AttachmentInfo {
	matches := attachmentRE.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return nil
	}
	out := make([]AttachmentInfo, 0, len(matches))
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		out = append(out, AttachmentInfo{
			Name: strings.TrimSpace(m[1]),
			Path: strings.TrimSpace(m[2]),
		})
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

// ReplaceNoteText replaces user-authored text while preserving attachment lines.
func ReplaceNoteText(raw, newText string) string {
	newText = strings.TrimSpace(newText)
	attachments := ParseAttachments(raw)
	if len(attachments) == 0 {
		if newText == "" {
			return raw
		}
		if !strings.HasSuffix(newText, "\n") {
			newText += "\n"
		}
		return newText
	}
	var lines []string
	if newText != "" {
		lines = append(lines, newText)
	}
	for _, att := range attachments {
		lines = append(lines, FormatAttachmentContent(att.Path, att.Name))
	}
	return strings.Join(lines, "\n") + "\n"
}

// ApplyDraftTitle prepends a user title and labels unnamed attachment links.
func ApplyDraftTitle(content, title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return content
	}
	content = attachmentRE.ReplaceAllStringFunc(content, func(match string) string {
		m := attachmentRE.FindStringSubmatch(match)
		if len(m) < 3 {
			return match
		}
		label := strings.TrimSpace(m[1])
		if label != "" && label != attachmentPlaceholder {
			return match
		}
		return fmt.Sprintf("📎 [%s](%s)", title, m[2])
	})
	if DraftTitle(content) == title {
		return content
	}
	return title + "\n" + strings.TrimSpace(content)
}
