package worker

import (
	"fmt"
	"path"
	"strconv"
	"time"

	"github.com/spf13/afero"
	"golang.org/x/exp/slog"

	"zakirullin/stuffbot/internal"
	"zakirullin/stuffbot/internal/db"
	"zakirullin/stuffbot/internal/fs"
	"zakirullin/stuffbot/internal/sched"
	"zakirullin/stuffbot/internal/userconfig"
	"zakirullin/stuffbot/pkg/txt"
)

const (
	daysInAdvanceForLater = 7 * 24 * time.Hour
)

func MoveDueTasksToToday(
	storagePath,
	configFilename string,
	fsBackend afero.Fs,
	telegram internal.TGInterface,
) error {
	rootFS, err := fs.NewFS(storagePath, fsBackend)
	if err != nil {
		return fmt.Errorf("schedule worker: can't create FS: %s", err)
	}

	userDirs, err := rootFS.FilesAndDirs(fs.DirRoot)
	if err != nil {
		return fmt.Errorf("schedule worker: %w", err)
	}
	// TODO release
	// userDirs = fs.OnlyUserDirs(userDirs)

	for _, userDir := range userDirs {
		userID, err := strconv.ParseInt(userDir.Name, 10, 64)
		if err != nil {
			return fmt.Errorf("schedule worker: can't parse user ID: %s", err)
		}
		userPath := path.Join(storagePath, txt.I64(userID))
		userFS, err := fs.NewFS(userPath, fsBackend)
		if err != nil {
			return fmt.Errorf("schedule worker: can't create FS: %s", err)
		}

		userconf := userconfig.NewConfig()
		userconfPath := userFS.UnsafePath("", configFilename)
		err = userconf.LoadOrCreate(userconfPath)
		if err != nil {
			return fmt.Errorf("schedule worker: can't load user config: %s", err)
		}

		for _, schedule := range userconf.Schedules() {
			secondsLeft := schedule.ScheduledAt - time.Now().Unix()
			shouldScheduleForToday := secondsLeft <= 0
			shouldScheduleForLater := secondsLeft > 0 && secondsLeft <= int64(daysInAdvanceForLater.Seconds())
			shouldNotSchedule := !shouldScheduleForToday && !shouldScheduleForLater
			if shouldNotSchedule {
				continue
			}

			if shouldScheduleForLater {
				err = moveTaskToLater(schedule.Filename, userFS)
				if err != nil {
					slog.Error("schedule worker: can't move to later", "err", err)
				}
				// We don't need to save the schedule, since we didn't modify it
				continue
			}

			err = moveTaskToToday(schedule.Filename, userFS)
			if err != nil {
				slog.Error("schedule worker: can't move to today", "err", err)
				continue
			}

			// Removing the task from the schedule
			bot := internal.NewBot(userID, telegram, userFS, db.NewDB(), userconf)
			_ = bot.ShowTodayTasks(nil)

			slog.Debug("Scheduled task moved to today", schedule.Filename, "filename")

			err = userconf.LoadOrCreate(userconfPath)
			if err != nil {
				return fmt.Errorf("schedule worker: can't load user config before save: %s", err)
			}
			userconf.DelFromSchedule(schedule.Filename)

			// Schedule a recurring task if cron is not empty
			if len(schedule.Cron) != 0 {
				scheduledAt := sched.Next(schedule.Cron)
				userconf.AddToSchedule(schedule.Filename, scheduledAt, schedule.Cron)
				slog.Debug("Task was rescheduled", "filename", schedule.Filename, "schedule", schedule.Cron, "scheduledAt", scheduledAt)
				continue
			}

			err = userconf.Save(userconfPath)
			if err != nil {
				return fmt.Errorf("schedule worker: can't save user config: %s", err)
			}
		}
	}

	return nil
}

func moveTaskToLater(filename string, userFS *fs.FS) error {
	filenames, err := userFS.FilesAndDirs(fs.DirArchive)
	if err != nil {
		return fmt.Errorf("moveTaskToLater: %w", err)
	}

	for _, f := range filenames {
		if f.Name == filename {
			err = userFS.Rename(fs.DirArchive, filename, fs.DirLater, filename)
			if err != nil {
				return fmt.Errorf("moveTaskToLater: can't rename: %w", err)
			}
		}
	}

	return nil
}

func moveTaskToToday(filename string, userFS *fs.FS) error {
	dirsToLookFor := []string{fs.DirLater, fs.DirArchive}
	for _, dir := range dirsToLookFor {
		filenames, err := userFS.FilesAndDirs(dir)
		if err != nil {
			return fmt.Errorf("moveTaskForToday: %w", err)
		}

		for _, f := range filenames {
			if f.Name == filename {
				err = userFS.Rename(dir, filename, fs.DirToday, filename)
				if err != nil {
					return fmt.Errorf("moveTaskForToday: can't rename: %w", err)
				}
			}
		}
	}

	return nil
}
