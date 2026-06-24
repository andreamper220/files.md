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

// FormatAttachmentContent returns markdown for a saved file without a user title.
func FormatAttachmentContent(mediaPath string) string {
	return fmt.Sprintf("📎 [%s](%s)", attachmentPlaceholder, mediaPath)
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
	return DraftTitle(raw) == ""
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

// ApplyDraftTitle prepends a user title and labels attachment links.
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
		return fmt.Sprintf("📎 [%s](%s)", title, m[2])
	})
	if DraftTitle(content) == title {
		return content
	}
	return title + "\n" + strings.TrimSpace(content)
}
