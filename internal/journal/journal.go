package journal

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"zakirullin/stuffbot/internal/fs"
	"zakirullin/stuffbot/internal/habits"
	"zakirullin/stuffbot/pkg/txt"
)

var now = time.Now

// AddRecord adds a record for the current day
// Creates a file if there's no one for the current month
func AddRecord(userFS *fs.FS, noteFilename string) error {
	record, err := userFS.Read(fs.DirToday, noteFilename)
	if err != nil {
		return fmt.Errorf("failed to move to journal: can't get note content: %w", err)
	}

	record = strings.TrimSpace(record)
	if len(record) == 0 {
		record = fs.Title(noteFilename)
	}

	journalFilename := todayJournalFilename()
	exists, err := userFS.Exists(fs.DirJournal, journalFilename)
	if err != nil {
		return err
	}

	var md string
	if exists {
		md, err = userFS.Read(fs.DirJournal, journalFilename)
		if err != nil {
			return err
		}
		md = txt.NormNewLines(md)
		md = strings.TrimSpace(md)
		if len(md) != 0 {
			md += "\n"
		}
	}

	if !strings.Contains(md, todayHeader()) {
		md += todayHeader() + "\n"
	}

	pattern := `(!\[\[.*?\]\]\s+)(.*)`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(record)
	if len(matches) > 2 {
		modifiedText := fmt.Sprintf("%s%s ", matches[1], now().Format("`15:04`"))
		record = strings.Replace(record, matches[1], modifiedText, 1)
		record = fmt.Sprintf("%s\n", strings.TrimSpace(record))
	} else {
		record = fmt.Sprintf("%s %s\n", now().Format("`15:04`"), record)
	}

	md += record

	return userFS.Write(fs.DirJournal, journalFilename, md)
}

// AddEmoji adds an emoji to the current day's record
// Creates a file if there's no one for the current month
func AddEmoji(userFS *fs.FS, emoji string) error {
	if len(emoji) == 0 {
		return nil
	}

	journalFilename := todayJournalFilename()
	exists, err := userFS.Exists(fs.DirJournal, journalFilename)
	if err != nil {
		return err
	}

	if !exists {
		md := fmt.Sprintf("%s %s", todayHeader(), emoji)
		return userFS.Write(fs.DirJournal, journalFilename, md)
	}

	md, err := userFS.Read(fs.DirJournal, journalFilename)
	if err != nil {
		return err
	}
	md = txt.NormNewLines(md)
	md = strings.TrimSpace(md)

	todayHeaderRE := regexp.MustCompile(fmt.Sprintf(`(%s) *(.*)`, todayHeader()))
	if todayHeaderRE.MatchString(md) {
		replacement := fmt.Sprintf(`$1 ${2}%s`, emoji)
		// Prepend day's mood emoji in front of all other emojis
		if slices.Contains(habits.MoodEmojis, emoji) {
			replacement = fmt.Sprintf(`$1 %s${2}`, emoji)
		}
		md = todayHeaderRE.ReplaceAllString(md, replacement)
	} else {
		md += fmt.Sprintf("\n%s %s", todayHeader(), emoji)
	}

	return userFS.Write(fs.DirJournal, journalFilename, md)
}

func todayJournalFilename() string {
	return now().Format("2006.01 January.md")
}

func todayHeader() string {
	return fmt.Sprintf("#### %d %s, %s", now().Day(), now().Format("January"), now().Weekday())
}
