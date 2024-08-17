// Package userconfig stores user's configuration in file.
// It stores such settings for users as: language, home, quick buttons, schedule and so on.
package userconfig

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"time"

	"golang.org/x/exp/slog"

	"zakirullin/stuffbot/internal/consts"
	"zakirullin/stuffbot/internal/fs"
)

var DefaultConfig = Config{ // TODO apply default config if some fields are missing
	raw: raw{
		Language: "en",
		HomeCmd:  "today",
		MoveToCmds: []string{
			consts.CmdScheduleForTmrw,
			consts.CmdLater,
			consts.CmdShowScheduleForDay,
			consts.CmdShowMoveToFile,
			consts.CmdMoveToJournal,
			//"checklist",
		},
		PomodoroDurationMinute: 25,
		Schedules:              []Schedule{},
		JournalFilenameFormat:  "January 2006.md",
		JournalHeaderFormat:    "02, Monday",
		QuickCmds:              []string{},
	},
}

type Config struct {
	raw
}

type Schedule struct {
	Filename    string
	ScheduledAt int64
	Cron        string
	Cmd         string // For future use
}

type raw struct {
	Language               string     `json:"language"`
	HomeCmd                string     `json:"homeCommand"`
	MoveToCmds             []string   `json:"moveToCommands"`
	PomodoroDurationMinute float64    `json:"pomodoroDurationMinute"`
	JournalFilenameFormat  string     `json:"journalFilename"`
	JournalHeaderFormat    string     `json:"journalHeaderFormat"`
	Schedules              []Schedule `json:"schedules"`
	QuickCmds              []string   `json:"quickCommands"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) LoadOrCreate(path string) error {
	exists, err := fs.Exists(path)
	if err != nil {
		return fmt.Errorf("config load: %w", err)
	}

	if !exists {
		c.raw = DefaultConfig.raw
		return nil
	}

	configFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("config load: %w", err)
	}
	defer configFile.Close()

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		return fmt.Errorf("config load: %w", err)
	}

	err = json.Unmarshal(bytes, c)
	if err != nil {
		return fmt.Errorf("config load: can't unmarshal: %w", err)
	}

	return nil
}

func (c *Config) Save(path string) error { // TODO add lazy saving, save only if config was changed
	bytes, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return fmt.Errorf("config save: can't marshal config: %w", err)
	}

	err = fs.WriteFile(path, bytes)
	if err != nil {
		return fmt.Errorf("config save: can't write config file: %w", err)
	}

	return nil
}

func (c *Config) SetPomodoroDuration(value time.Duration) error {
	if value <= 0 || value > 24*time.Hour {
		return fmt.Errorf("raw.SetPomodoroDuration: value is invalid: %v", value)
	}
	c.raw.PomodoroDurationMinute = value.Minutes()
	return nil
}

func (c *Config) PomodoroDuration() time.Duration {
	minutes := c.raw.PomodoroDurationMinute
	if minutes <= 0 {
		slog.Error("Pomodoro duration is invalid. Using default value", "duration",
			c.raw.PomodoroDurationMinute, "default", DefaultConfig.raw.PomodoroDurationMinute)
		// I don't use DefaultConfig.PomodoroDuration() because it may cause infinite recursion
		minutes = DefaultConfig.raw.PomodoroDurationMinute
	}
	return time.Duration(minutes * float64(time.Minute))
}

func (c *Config) Schedules() []Schedule {
	schedules := c.raw.Schedules
	sort.Slice(c.raw.Schedules, func(i, j int) bool {
		return c.raw.Schedules[i].ScheduledAt > c.raw.Schedules[j].ScheduledAt
	})
	slices.Reverse(schedules)

	return schedules
}

// AddToSchedule task from archive or later at scheduleAt (Unix timestamp, sec). Tasks appear in today folder.
// If cron is provided this task will be repeated. Other wise, it will be executed once.
func (c *Config) AddToSchedule(filename string, scheduleAt int64, cron string) {
	c.raw.Schedules = append(c.raw.Schedules, Schedule{filename, scheduleAt, cron, ""})
}

func (c *Config) DelFromSchedule(filename string) {
	var newSchedules []Schedule
	for _, schedule := range c.raw.Schedules {
		if schedule.Filename != filename {
			newSchedules = append(newSchedules, schedule)
		}
	}
	c.raw.Schedules = newSchedules
}

func (c *Config) JournalFilenameFormat() string {
	if c.raw.JournalFilenameFormat == "" {
		return DefaultConfig.raw.JournalFilenameFormat
	}
	return c.raw.JournalFilenameFormat
}

func (c *Config) SetJournalFilenameFormat(path string) {
	c.raw.JournalFilenameFormat = path
}

func (c *Config) JournalHeaderFormat() string {
	if c.raw.JournalHeaderFormat == "" {
		return DefaultConfig.raw.JournalHeaderFormat
	}
	return c.raw.JournalHeaderFormat
}

func (c *Config) SetJournalHeaderFormat(format string) {
	c.raw.JournalHeaderFormat = format
}
