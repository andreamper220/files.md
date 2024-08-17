package sched

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"

	"zakirullin/stuffbot/internal/fs"
	"zakirullin/stuffbot/internal/userconfig"
)

var now = func() time.Time {
	return time.Now()
}

type Cron struct {
	RunAt int64
	Cron  string
	Cmd   string // For future use
}

func NewCron(runAt int64, cron string) Cron {
	return Cron{runAt, cron, "move"}
}

func BeginningOfTheDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func Tomorrow() int64 {
	tomorrow := now().AddDate(0, 0, 1)

	return BeginningOfTheDay(tomorrow).Unix()
}

// Next returns next unix time for cron expression
func Next(crn string) int64 {
	sched, err := cron.ParseStandard(crn)
	// TODO release, we should not panic when a user provided bad config
	if err != nil {
		// It's a logical error in code, we don't obtain cron expressions from user input
		panic(fmt.Errorf("invalid cron expression %s: %w", crn, err))
	}

	return sched.Next(now().UTC()).Unix()
}

func ScheduleReport(conf *userconfig.Config) string {
	var report string
	scheduledTasks := conf.Schedules()
	for _, task := range scheduledTasks {
		report += fmt.Sprintf("<b>%s</b>: %s\n", formatTaskDate(task.ScheduledAt), fs.Title(task.Filename))
	}

	return report
}

// TODO write tests for that
func formatTaskDate(scheduledAt int64) string {
	today := now().Truncate(24 * time.Hour)
	taskDate := time.Unix(scheduledAt, 0).Truncate(24 * time.Hour)

	diffDays := int(taskDate.Sub(today).Hours() / 24)

	switch {
	case diffDays == 0:
		return "Today"
	case diffDays == 1:
		return "Tomorrow"
	case diffDays > 1 && diffDays <= 6: // Nearest day
		return taskDate.Weekday().String()
	case diffDays >= 7 && diffDays <= 13:
		return "Next " + taskDate.Weekday().String()
	default:
		return taskDate.Format("02 January")
	}
}
