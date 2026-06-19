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

// Build returns emoji-only home summary: spheres → areas with done-task and new-note counts.
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
		lines = append(lines, life.SphereEmoji(spherePath))
		projects, err := life.ListProjects(userFS, spherePath)
		if err != nil {
			continue
		}
		for _, projectPath := range projects {
			line := " " + life.AreaEmoji(projectPath)
			for _, emoji := range emojis {
				n := countDoneTasksByPriority(userFS, projectPath, emoji, emojis)
				if n > 0 {
					line += fmt.Sprintf(" %s%d", emoji, n)
				}
			}
			for _, kind := range []life.Kind{life.KindDraft, life.KindFinal, life.KindDiscussion} {
				n := countNewNotes(userFS, projectPath, kind, startOfDay)
				if n > 0 {
					line += fmt.Sprintf(" %s%d", life.KindEmoji(kind), n)
				}
			}
			lines = append(lines, line)
		}
	}

	// Inbox tasks without area
	inboxDone := countInboxDoneByPriority(userFS, emojis, startOfDay)
	if len(inboxDone) > 0 {
		lines = append(lines, "📥")
		line := " ⚪️"
		for _, emoji := range emojis {
			if n := inboxDone[emoji]; n > 0 {
				line += fmt.Sprintf(" %s%d", emoji, n)
			}
		}
		lines = append(lines, strings.TrimSpace(line))
	}

	if len(lines) == 0 {
		return "", nil
	}
	return strings.Join(lines, "\n"), nil
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
		lines = append(lines, life.SphereEmoji(spherePath))
		projects, _ := life.ListProjects(userFS, spherePath)
		for _, projectPath := range projects {
			parts := []string{" " + life.AreaEmoji(projectPath)}
			for _, kind := range []life.Kind{life.KindDraft, life.KindFinal, life.KindDiscussion} {
				total := countNotesInDir(userFS, life.DocDir(projectPath, kind))
				today := countNewNotes(userFS, projectPath, kind, startOfDay)
				if total > 0 || today > 0 {
					parts = append(parts, fmt.Sprintf("%s %d", life.KindEmoji(kind), today))
				}
			}
			if len(parts) > 1 {
				lines = append(lines, strings.Join(parts, " "))
			}
		}
	}

	if len(lines) == 1 {
		lines = append(lines, i18nNothingToday())
	}
	return strings.Join(lines, "\n"), nil
}

func i18nNothingToday() string {
	return "—"
}

func countDoneTasksByPriority(userFS *fs.FS, projectPath, emoji string, emojis []string) int {
	md, err := userFS.Read(projectPath, life.TasksFilename)
	if err != nil {
		return 0
	}
	return countDoneInMD(md, emoji, emojis)
}

func countInboxDoneByPriority(userFS *fs.FS, emojis []string, since time.Time) map[string]int {
	out := map[string]int{}
	md, err := userFS.Read(fs.DirUserRoot, fs.ChatFilename)
	if err != nil {
		return out
	}
	for _, task := range parseTasks(md) {
		if !task.done {
			continue
		}
		e := priority.Detect(task.text, emojis)
		if e == "" {
			e = "⚪️"
		}
		out[e]++
	}
	_ = since
	return out
}

func countDoneInMD(md, emoji string, emojis []string) int {
	n := 0
	for _, task := range parseTasks(md) {
		if !task.done {
			continue
		}
		e := priority.Detect(task.text, emojis)
		if e == "" {
			e = "⚪️"
		}
		if e == emoji {
			n++
		}
	}
	return n
}

func countNewNotes(userFS *fs.FS, projectPath string, kind life.Kind, since time.Time) int {
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

func countNotesInDir(userFS *fs.FS, dir string) int {
	entries, err := userFS.FilesAndDirs(dir)
	if err != nil {
		return 0
	}
	return len(fs.OnlyFiles(entries))
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
