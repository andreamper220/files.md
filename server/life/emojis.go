package life

import (
	"fmt"
	"hash/fnv"
	"strings"
)

// Sphere emoji markers for home summary (names are not shown).
var sphereEmojis = map[string]string{
	"Работа":      "💼",
	"Work":        "💼",
	"Отношения":   "❤️",
	"Relationships": "❤️",
	"Здоровье":    "💚",
	"Health":      "💚",
	"Личное":      "🌿",
	"Personal":    "🌿",
	"Обучение":    "📚",
	"Learning":    "📚",
}

var areaEmojiPool = []string{"🏗", "📦", "🎯", "🔧", "📌", "🧩", "🛠", "📎", "🗂", "🧪"}

// TasksFilename stores per-area task checklists.
const TasksFilename = "tasks.md"

// SphereEmoji returns an emoji for a sphere path (no text label).
func SphereEmoji(spherePath string) string {
	title := SphereTitle(spherePath)
	if e, ok := sphereEmojis[title]; ok {
		return e
	}
	name := baseName(spherePath)
	if e, ok := sphereEmojis[name]; ok {
		return e
	}
	return pickEmoji(name, sphereEmojiPool())
}

// AreaEmoji returns an emoji for an area (project) path.
func AreaEmoji(projectPath string) string {
	name := baseName(projectPath)
	return pickEmoji(name, areaEmojiPool)
}

// KindEmoji returns an emoji for a document kind.
func KindEmoji(k Kind) string {
	switch k {
	case KindDraft:
		return "📝"
	case KindFinal:
		return "✨"
	case KindDiscussion:
		return "💬"
	default:
		return "📝"
	}
}

func sphereEmojiPool() []string {
	out := make([]string, 0, len(sphereEmojis))
	seen := map[string]bool{}
	for _, e := range sphereEmojis {
		if !seen[e] {
			seen[e] = true
			out = append(out, e)
		}
	}
	return out
}

func pickEmoji(seed string, pool []string) string {
	if len(pool) == 0 {
		return "⚪️"
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(seed))
	return pool[int(h.Sum32())%len(pool)]
}

// AreaListPrefix marks areas nested under spheres in lists.
const AreaListPrefix = "└ "

// AreaLabel returns emoji + title for UI buttons.
func AreaLabel(projectPath string) string {
	return AreaEmoji(projectPath) + " " + AreaTitle(projectPath)
}

// NestedAreaLabel returns an area label with a tree marker for nested lists.
func NestedAreaLabel(projectPath string) string {
	return AreaTreePrefix(AreaDepth(projectPath)) + AreaLabel(projectPath)
}

// AreaTreePrefix returns indentation for a nested area depth (1 = top-level).
func AreaTreePrefix(depth int) string {
	if depth <= 1 {
		return AreaListPrefix
	}
	return strings.Repeat("│  ", depth-1) + "└ "
}

// SaveLocationLabel returns sphere + area names for save confirmations.
func SaveLocationLabel(spherePath, areaPath string) string {
	return fmt.Sprintf("%s %s → %s %s",
		SphereEmoji(spherePath), SphereTitle(spherePath),
		AreaEmoji(areaPath), AreaFullTitle(areaPath))
}

// AreaPickerLabel returns a button label for area pickers.
func AreaPickerLabel(spherePath, areaPath string) string {
	return SaveLocationLabel(spherePath, areaPath)
}

// SphereLabel returns emoji-only label for UI buttons.
func SphereLabel(spherePath string) string {
	return SphereEmoji(spherePath)
}
