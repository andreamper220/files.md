package insights

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rivo/uniseg"

	"zakirullin/stuffbot/internal/fs"
	"zakirullin/stuffbot/pkg/txt"
)

// [1 => false, <year day> => false, ...]
type Year map[int]bool

const (
	habitSkipped            = "⚪️"
	habitCompleted          = "🟢"
	habitCompletedAtWeekend = "🟡"
)

var (
	errMalformedMonthLine = errors.New("malformed month line")
)

// getLastWeekHabits
// getLastMonthHabits

func ReadHabits(botFS *fs.FS, year int) (map[string]Year, error) {
	filename := "%d Habits.md"
	habitsStr, err := botFS.Read(fs.DirInsights, fmt.Sprintf(filename, year))
	if err != nil {
		return nil, fmt.Errorf("read %s error: %w", filename, err)
	}

	habits := make(map[string]Year)
	month := time.January
	lines := strings.Split(txt.NormNewLines(habitsStr), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Parsing month line
		isMonthLine := strings.HasPrefix(line, "###")
		if isMonthLine {
			parts := strings.Split(line, " ")
			if len(parts) < 2 {
				return nil, fmt.Errorf("read habits: can't parse month line '%s': %w", line, errMalformedMonthLine)
			}

			date, err := time.Parse("January", parts[1])
			if err != nil {
				return nil, fmt.Errorf("read habits: can't parse month %s: %w", line, err)
			}
			month = date.Month()

			continue
		}

		// Tolerant reader, if we encounter gibberish,
		// we skip it. See ADRs in README.md for details for details
		isHabbitLine := strings.ContainsAny(line, fmt.Sprintf("%s%s", habitSkipped, habitCompleted))
		if !isHabbitLine {
			continue
		}

		// At this point we are on habits line, which is
		// [⚪️🟢... Habit name] i.e. completion status
		// for every day of the above found month

		daysAndHabit := strings.SplitN(line, " ", 2)
		if len(daysAndHabit) < 2 {
			return nil, nil
			// return "bad month line: %s"
		}
		habitName := strings.TrimSpace(daysAndHabit[1])
		if _, ok := habits[habitName]; !ok {
			habits[habitName] = make(Year)
		}

		firstDayOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		dayOfTheYear := firstDayOfMonth.YearDay()

		days := daysAndHabit[0]
		// See README.md ADRs
		gr := uniseg.NewGraphemes(days)
		dayOffset := 0
		for gr.Next() {
			habits[habitName][dayOfTheYear+dayOffset] = gr.Str() != habitSkipped
			dayOfTheYear++
		}
	}

	return habits, nil
}

// func Write(botFS *fs.FS, habits []Habit) error {
// 	return nil
// }
