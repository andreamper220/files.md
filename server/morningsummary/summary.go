package morningsummary

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/pkg/txt"
	"github.com/zakirullin/files.md/server/priority"
	"github.com/zakirullin/files.md/server/userconfig"
)

var chatTaskRE = regexp.MustCompile("^- \\[([ xX])\\] (?:`\\d{2}:\\d{2}` )?(.*)$")

type bucket struct {
	done  int
	total int
}

// Build returns an HTML morning report for today's tasks.
func Build(userFS *fs.FS, cfg *userconfig.Config) (string, error) {
	emojis := cfg.PriorityEmojis()
	categories := cfg.TaskCategories()

	type source struct {
		name string
		md   string
	}
	var sources []source

	chatMD, err := userFS.Read(fs.DirUserRoot, fs.ChatFilename)
	if err == nil {
		sources = append(sources, source{"Сегодня", chatMD})
	}
	laterMD, err := userFS.Read(fs.DirUserRoot, fs.LaterFilename)
	if err == nil {
		sources = append(sources, source{"Позже", laterMD})
	}
	for _, cat := range categories {
		filename := categoryFilename(cat)
		md, err := userFS.Read(fs.DirUserRoot, filename)
		if err != nil {
			continue
		}
		sources = append(sources, source{cat, md})
	}

	byCategory := map[string]bucket{}
	byPriority := map[string]bucket{}
	totalDone, totalAll := 0, 0

	for _, src := range sources {
		cb := byCategory[src.name]
		for _, task := range parseTasks(src.md) {
			raw := strings.TrimSpace(task.text)
			if raw == "" {
				continue
			}
			cb.total++
			totalAll++
			if task.done {
				cb.done++
				totalDone++
			}
			emoji := priority.Detect(raw, emojis)
			if emoji == "" {
				emoji = "—"
			}
			pb := byPriority[emoji]
			pb.total++
			if task.done {
				pb.done++
			}
			byPriority[emoji] = pb
		}
		byCategory[src.name] = cb
	}

	var lines []string
	lines = append(lines, "<b>🌅 Утренняя сводка</b>")
	lines = append(lines, fmt.Sprintf("Всего: <b>%d/%d</b> выполнено", totalDone, totalAll))
	lines = append(lines, "")

	if len(byCategory) > 0 {
		lines = append(lines, "<b>По категориям</b>")
		for _, src := range sources {
			cb := byCategory[src.name]
			if cb.total == 0 {
				continue
			}
			lines = append(lines, fmt.Sprintf("• %s: %d/%d", src.name, cb.done, cb.total))
		}
		lines = append(lines, "")
	}

	if len(byPriority) > 0 {
		lines = append(lines, "<b>По срочности</b>")
		order := append([]string{}, emojis...)
		order = append(order, "—")
		seen := map[string]bool{}
		for _, emoji := range order {
			cb, ok := byPriority[emoji]
			if !ok || cb.total == 0 {
				continue
			}
			seen[emoji] = true
			label := emoji
			if emoji == "—" {
				label = "без метки"
			}
			lines = append(lines, fmt.Sprintf("• %s: %d/%d", label, cb.done, cb.total))
		}
		for emoji, cb := range byPriority {
			if seen[emoji] || cb.total == 0 {
				continue
			}
			lines = append(lines, fmt.Sprintf("• %s: %d/%d", emoji, cb.done, cb.total))
		}
	}

	return strings.TrimSpace(strings.Join(lines, "\n")), nil
}

func categoryFilename(category string) string {
	return fs.SanitizeFilename(category) + "_.md"
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
