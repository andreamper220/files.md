package morningsummary

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/life"
	"github.com/zakirullin/files.md/server/pkg/txt"
	"github.com/zakirullin/files.md/server/priority"
	"github.com/zakirullin/files.md/server/userconfig"
)

var chatTaskRE = regexp.MustCompile("^- \\[([ xX])\\] (?:`\\d{2}:\\d{2}` )?(.*)$")

const noPriorityEmoji = "⚪️"

// displayPriorityEmojis returns urgency emojis shown on home (no "unlabeled" bucket).
func displayPriorityEmojis(emojis []string) []string {
	out := make([]string, 0, len(emojis))
	for _, e := range emojis {
		if e == noPriorityEmoji {
			continue
		}
		out = append(out, e)
	}
	return out
}

// Build returns home summary: spheres → areas with names, urgency counts (emoji only), and new-note totals.
func Build(userFS *fs.FS, cfg *userconfig.Config) (string, error) {
	return buildHomeSummary(userFS, cfg)
}

func buildHomeSummary(userFS *fs.FS, cfg *userconfig.Config) (string, error) {
	_ = life.EnsureSpheresRoot(userFS)
	spheres, err := life.ListSpheres(userFS)
	if err != nil {
		return "", err
	}

	emojis := cfg.PriorityEmojis()
	tz := cfg.Timezone()
	startOfDay := beginningOfDay(time.Now().In(tz))

	var lines []string
	for _, spherePath := range spheres {
		projects, err := life.ListAllAreas(userFS, spherePath)
		if err != nil {
			continue
		}

		if len(projects) == 0 {
			lines = append(lines, formatSphereOnlyLine(spherePath, emojis, userFS, startOfDay))
			continue
		}

		lines = append(lines, formatSphereHeader(spherePath))
		for _, projectPath := range projects {
			lines = append(lines, formatAreaBlock(projectPath, emojis, userFS, startOfDay)...)
		}
	}

	if len(lines) == 0 {
		return "", nil
	}
	return strings.Join(lines, "\n"), nil
}

// formatSphereHeader builds "💼 Работа".
func formatSphereHeader(spherePath string) string {
	return life.SphereEmoji(spherePath) + " " + life.SphereTitle(spherePath)
}

// formatAreaBlock builds tabulated area summary with counts under the title.
func formatAreaBlock(projectPath string, emojis []string, userFS *fs.FS, startOfDay time.Time) []string {
	depth := life.AreaDepth(projectPath)
	header := life.AreaTreePrefix(depth) + formatAreaHeader(projectPath)
	indent := strings.Repeat("   ", depth)
	taskLine := indent + strings.Join(formatPriorityCounts(emojis, countOpenByPriorityInMD(readAreaTasksMD(userFS, projectPath), emojis)), " ")
	noteLine := indent + formatNoteTotal(userFS, projectPath, startOfDay)
	return []string{header, taskLine, noteLine}
}

// formatAreaLine builds a single-line area summary (legacy helper).
func formatAreaLine(projectPath string, emojis []string, userFS *fs.FS, startOfDay time.Time) string {
	return strings.Join(formatAreaBlock(projectPath, emojis, userFS, startOfDay), "\n")
}

func formatAreaHeader(projectPath string) string {
	return life.AreaEmoji(projectPath) + " " + life.AreaFullTitle(projectPath)
}

func formatSphereOnlyLine(spherePath string, emojis []string, userFS *fs.FS, startOfDay time.Time) string {
	lines := []string{formatSphereHeader(spherePath)}
	lines = append(lines, "   "+strings.Join(formatPriorityCounts(emojis, emptyPriorityCounts(emojis)), " "))
	lines = append(lines, "   "+formatNoteTotalForSphere(spherePath, userFS, startOfDay))
	return strings.Join(lines, "\n")
}

func formatPriorityCounts(emojis []string, counts map[string]int) []string {
	var parts []string
	for _, emoji := range displayPriorityEmojis(emojis) {
		parts = append(parts, fmt.Sprintf("%s%d", emoji, counts[emoji]))
	}
	return parts
}

func formatNoteTotal(userFS *fs.FS, projectPath string, startOfDay time.Time) string {
	return fmt.Sprintf("📝 %d", totalNewNotes(userFS, projectPath, startOfDay))
}

func formatNoteTotalForSphere(spherePath string, userFS *fs.FS, startOfDay time.Time) string {
	return fmt.Sprintf("📝 %d", totalNewNotesForSphere(spherePath, userFS, startOfDay))
}

func totalNewNotes(userFS *fs.FS, projectPath string, startOfDay time.Time) int {
	if projectPath == "" {
		return 0
	}
	n := 0
	for _, kind := range []life.Kind{life.KindDraft, life.KindFinal, life.KindDiscussion} {
		n += countNewNotes(userFS, projectPath, kind, startOfDay)
	}
	return n
}

func totalNewNotesForSphere(spherePath string, userFS *fs.FS, startOfDay time.Time) int {
	projects, _ := life.ListAllAreas(userFS, spherePath)
	n := 0
	for _, projectPath := range projects {
		n += totalNewNotes(userFS, projectPath, startOfDay)
	}
	return n
}

func emptyPriorityCounts(emojis []string) map[string]int {
	counts := map[string]int{}
	for _, emoji := range displayPriorityEmojis(emojis) {
		counts[emoji] = 0
	}
	return counts
}

func mergePriorityCounts(dst, src map[string]int) {
	for k, v := range src {
		dst[k] += v
	}
}

func readAreaTasksMD(userFS *fs.FS, projectPath string) string {
	md, err := userFS.Read(projectPath, life.TasksFilename)
	if err != nil {
		return ""
	}
	return md
}

func countOpenByPriorityInMD(md string, emojis []string) map[string]int {
	counts := emptyPriorityCounts(emojis)
	for _, task := range parseTasks(md) {
		if task.done {
			continue
		}
		emoji := priority.Detect(task.text, emojis)
		if emoji == "" || emoji == noPriorityEmoji {
			continue
		}
		counts[emoji]++
	}
	return counts
}

// BuildNotesHub returns detailed notes statistics for the notes screen.
func BuildNotesHub(userFS *fs.FS, cfg *userconfig.Config) (string, error) {
	_ = life.EnsureSpheresRoot(userFS)
	tz := cfg.Timezone()
	startOfDay := beginningOfDay(time.Now().In(tz))

	var lines []string
	lines = append(lines, "<b>🗒</b>")

	spheres, _ := life.ListSpheres(userFS)
	for _, spherePath := range spheres {
		lines = append(lines, formatSphereHeader(spherePath))
		areas, _ := life.ListAllAreas(userFS, spherePath)
		for _, projectPath := range areas {
			depth := life.AreaDepth(projectPath)
			header := life.AreaTreePrefix(depth) + formatAreaHeader(projectPath)
			indent := strings.Repeat("   ", depth)
			lines = append(lines, header)
			lines = append(lines, indent+fmt.Sprintf("📝 %d", totalNewNotes(userFS, projectPath, startOfDay)))
		}
	}

	if len(lines) == 1 {
		lines = append(lines, "—")
	}
	return strings.Join(lines, "\n"), nil
}

func countNewNotes(userFS *fs.FS, projectPath string, kind life.Kind, since time.Time) int {
	if projectPath == "" {
		return 0
	}
	dir := life.DocDir(projectPath, kind)
	entries, err := userFS.FilesAndDirs(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, f := range fs.OnlyFiles(entries) {
		if !strings.HasSuffix(f.Name, fs.MDExt) {
			continue
		}
		if fileTime(f.Ctime).In(since.Location()).Before(since) {
			continue
		}
		n++
	}
	return n
}

func beginningOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

type parsedTask struct {
	text string
	done bool
}

func parseTasks(md string) []parsedTask {
	md = txt.NormNewLines(md)
	var tasks []parsedTask
	for _, line := range strings.Split(md, "\n") {
		line = strings.TrimSpace(line)
		m := chatTaskRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		tasks = append(tasks, parsedTask{
			text: m[2],
			done: m[1] == "x" || m[1] == "X",
		})
	}
	return tasks
}
