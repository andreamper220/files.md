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
	"sync"
	"time"

	"zakirullin/stuffbot/internal/consts"
	"zakirullin/stuffbot/internal/fs"
)

var DefaultConfig = Config{ // TODO apply default config if some fields are missing

}

var defaultConfig = config{
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
	PomodoroDurationInMinutes: 50,
	Schedules:                 []Schedule{},
	QuickCmds:                 []string{},
}

var (
	mu        sync.Mutex
	userLocks map[int64]*sync.Mutex
)

type Config struct {
	userID int64
	path   string
}

type Schedule struct {
	Filename    string
	ScheduledAt int64
	Cron        string
	Cmd         string // For future use
}

type config struct {
	Language                  string     `json:"language"`
	HomeCmd                   string     `json:"homeCommand"`
	MoveToCmds                []string   `json:"moveToCommands"`
	PomodoroDurationInMinutes float64    `json:"pomodoroDurationInMinutes"`
	Schedules                 []Schedule `json:"schedules"`
	QuickCmds                 []string   `json:"quickCommands"`
}

func NewConfig(userID int64, path string) *Config {
	return &Config{userID: userID, path: path}
}

func (c *Config) CreateDefaultIfNotExists() error {
	exists, err := fs.Exists(c.path)
	if err != nil {
		return fmt.Errorf("can't check whether config exists: %w", err)
	}
	if exists {
		return nil
	}

	bytes, err := json.MarshalIndent(defaultConfig, "", "    ")
	if err != nil {
		return fmt.Errorf("can't marshal default config: %w", err)
	}

	err = fs.WriteFile(c.path, bytes)
	if err != nil {
		return fmt.Errorf("can't write default config file: %w", err)
	}

	return nil
}

func (c *Config) SetPomodoroDuration(duration time.Duration) error {
	if duration <= 0 || duration > 24*time.Hour {
		return fmt.Errorf("set pomodoro duration: duration is invalid: %v", duration)
	}

	lock := c.userLock()
	lock.Lock()
	defer lock.Unlock()

	conf, err := c.read(c.path)
	if err != nil {
		return fmt.Errorf("set pomodoro duration: can't read config: %w", err)
	}
	conf.PomodoroDurationInMinutes = duration.Minutes()
	err = c.write(c.path, conf)
	if err != nil {
		return fmt.Errorf("set pomodoro duration: can't write config: %w", err)
	}

	return nil
}

func (c *Config) PomodoroDuration() time.Duration {
	conf, _ := c.read(c.path)

	return time.Duration(conf.PomodoroDurationInMinutes * float64(time.Minute))
}

func (c *Config) Schedules() ([]Schedule, error) {
	conf, err := c.read(c.path)
	if err != nil {
		return nil, fmt.Errorf("can't get schedules: can't read config: %w", err)
	}

	schedules := conf.Schedules
	sort.Slice(schedules, func(i, j int) bool {
		return schedules[i].ScheduledAt > schedules[j].ScheduledAt
	})
	slices.Reverse(schedules)

	return schedules, nil
}

func (c *Config) AddToSchedule(filename string, scheduleAt int64, cron string) error {
	lock := c.userLock()
	lock.Lock()
	defer lock.Unlock()

	conf, err := c.read(c.path)
	if err != nil {
		return fmt.Errorf("can't add to schedule: can't read config: %w", err)
	}
	conf.Schedules = append(conf.Schedules, Schedule{filename, scheduleAt, cron, ""})
	err = c.write(c.path, conf)
	if err != nil {
		return fmt.Errorf("can't add to schedule: can't write config: %w", err)
	}

	return nil
}

func (c *Config) DelFromSchedule(filename string, scheduledAt int64) error {
	lock := c.userLock()
	lock.Lock()
	defer lock.Unlock()

	conf, err := c.read(c.path)
	if err != nil {
		return fmt.Errorf("can't del from schedule: can't read config: %w", err)
	}

	var newSchedules []Schedule
	for _, schedule := range conf.Schedules {
		if schedule.Filename == filename && schedule.ScheduledAt == scheduledAt {
			continue
		}
		newSchedules = append(newSchedules, schedule)
	}
	conf.Schedules = newSchedules

	err = c.write(c.path, conf)
	if err != nil {
		return fmt.Errorf("can't del from schedule: can't write config: %w", err)
	}

	return nil
}

func (c *Config) ShouldSplitChecklist(checklist string) bool {
	for _, unsplittableChecklist := range []string{fs.DirRead, fs.DirWatch} {
		if checklist == unsplittableChecklist {
			return false
		}
	}
	return true
}

func (c *Config) read(path string) (config, error) {
	exists, err := fs.Exists(path)
	if err != nil {
		return defaultConfig, fmt.Errorf("config load: %w", err)
	}

	if !exists {
		return defaultConfig, nil
	}

	configFile, err := os.Open(path)
	if err != nil {
		return defaultConfig, fmt.Errorf("config load: %w", err)
	}
	defer configFile.Close()

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		return defaultConfig, fmt.Errorf("config load: %w", err)
	}

	conf := config{}
	err = json.Unmarshal(bytes, &conf)
	if err != nil {
		return defaultConfig, fmt.Errorf("config load: can't unmarshal: %w", err)
	}

	return conf, nil
}

func (c *Config) write(path string, conf config) error {
	bytes, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		return fmt.Errorf("config save: can't marshal config: %w", err)
	}

	err = fs.WriteFile(path, bytes)
	if err != nil {
		return fmt.Errorf("config save: can't write config file: %w", err)
	}

	return nil
}

func (c *Config) userLock() *sync.Mutex {
	mu.Lock()
	defer mu.Unlock()

	if lock, exists := userLocks[c.userID]; exists {
		return lock
	}

	// TODO fix real id
	newLock := &sync.Mutex{}
	userLocks[1] = newLock

	return newLock

}
