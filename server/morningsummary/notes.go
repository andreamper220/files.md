package morningsummary

import (
	"fmt"
	"sort"
	"time"

	"github.com/zakirullin/files.md/server/fs"
)

const maxNoteLines = 5

type noteActivity struct {
	path string
	ts   int64
}

func buildNotesSection(userFS *fs.FS, tz *time.Location) []string {
	notes, err := userFS.AllNoteFiles()
	if err != nil || len(notes) == 0 {
		if len(notes) == 0 {
			return []string{
				"<b>📝 Заметки</b>",
				"Всего: <b>0</b>",
				"Сегодня без изменений",
			}
		}
		return nil
	}

	now := time.Now().In(tz)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, tz)

	var addedToday, modifiedToday []noteActivity
	for _, note := range notes {
		ctime := fileTime(note.Ctime).In(tz)
		mtime := fileTime(note.Mtime).In(tz)
		path := note.DisplayPath()

		if !ctime.Before(startOfDay) {
			addedToday = append(addedToday, noteActivity{path: path, ts: note.Ctime})
			continue
		}
		if !mtime.Before(startOfDay) {
			modifiedToday = append(modifiedToday, noteActivity{path: path, ts: note.Mtime})
		}
	}

	sort.Slice(addedToday, func(i, j int) bool { return addedToday[i].ts > addedToday[j].ts })
	sort.Slice(modifiedToday, func(i, j int) bool { return modifiedToday[i].ts > modifiedToday[j].ts })

	var lines []string
	lines = append(lines, "<b>📝 Заметки</b>")
	lines = append(lines, fmt.Sprintf("Всего: <b>%d</b>", len(notes)))

	if len(addedToday) > 0 {
		lines = append(lines, fmt.Sprintf("Добавлено сегодня: <b>%d</b>", len(addedToday)))
		lines = append(lines, formatNoteLines(addedToday, tz)...)
	}
	if len(modifiedToday) > 0 {
		lines = append(lines, fmt.Sprintf("Изменено сегодня: <b>%d</b>", len(modifiedToday)))
		lines = append(lines, formatNoteLines(modifiedToday, tz)...)
	}
	if len(addedToday) == 0 && len(modifiedToday) == 0 {
		lines = append(lines, "Сегодня без изменений")
		recent := recentNoteActivity(notes, tz, 7*24*time.Hour, maxNoteLines)
		if len(recent) > 0 {
			lines = append(lines, "Недавно:")
			lines = append(lines, formatNoteLines(recent, tz)...)
		}
	}

	return lines
}

func recentNoteActivity(notes []fs.NoteFile, tz *time.Location, within time.Duration, limit int) []noteActivity {
	cutoff := time.Now().In(tz).Add(-within)
	var recent []noteActivity
	for _, note := range notes {
		ts := note.Mtime
		if ts < note.Ctime {
			ts = note.Ctime
		}
		if fileTime(ts).In(tz).Before(cutoff) {
			continue
		}
		recent = append(recent, noteActivity{path: note.DisplayPath(), ts: ts})
	}
	sort.Slice(recent, func(i, j int) bool { return recent[i].ts > recent[j].ts })
	if len(recent) > limit {
		recent = recent[:limit]
	}
	return recent
}

func formatNoteLines(items []noteActivity, tz *time.Location) []string {
	limit := len(items)
	if limit > maxNoteLines {
		limit = maxNoteLines
	}
	var lines []string
	for _, item := range items[:limit] {
		lines = append(lines, fmt.Sprintf("• %s — %s", formatNoteTime(item.ts, tz), item.path))
	}
	if len(items) > maxNoteLines {
		lines = append(lines, fmt.Sprintf("• …ещё %d", len(items)-maxNoteLines))
	}
	return lines
}

func formatNoteTime(ts int64, tz *time.Location) string {
	t := fileTime(ts).In(tz)
	now := time.Now().In(tz)
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return t.Format("15:04")
	}
	return t.Format("02.01 15:04")
}

func fileTime(ts int64) time.Time {
	if ts > 1e12 {
		return time.UnixMicro(ts)
	}
	return time.Unix(ts, 0)
}
