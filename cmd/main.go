package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
	"github.com/spf13/afero"
	"golang.org/x/exp/slog"

	"zakirullin/stuffbot/config"
	"zakirullin/stuffbot/i18n"
	"zakirullin/stuffbot/internal"
	"zakirullin/stuffbot/internal/db"
	"zakirullin/stuffbot/internal/fs"
	"zakirullin/stuffbot/internal/sched/worker"
	"zakirullin/stuffbot/internal/sync"
	"zakirullin/stuffbot/internal/userconfig"
	"zakirullin/stuffbot/pkg/tg"
	"zakirullin/stuffbot/pkg/txt"
)

func main() {
	opts := &tint.Options{
		Level: slog.LevelDebug,
	}
	logger := slog.New(tint.NewHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	err := godotenv.Load()
	if err != nil {
		panic(fmt.Sprintf("Error loading .env file: %s\n", err))
	}
	err = config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Error loading conf: %s\n", err))
	}

	// TODO move to embed
	err = i18n.LoadLangFile("i18n/ru.json")
	if err != nil {
		panic(fmt.Sprintf("Error loading i18n: %s\n", err))
	}

	api, err := tgbotapi.NewBotAPI(config.Config.BotAPIToken)
	if err != nil {
		panic(fmt.Sprintf("Can't create TG api: %s\n", err))
	}
	telegram := tg.NewTG(api)

	// Initiate per-user locker
	userLocker := sync.NewPerUserLocker()

	// Workers
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	defer func(quit chan struct{}) {
		close(quit)
	}(quit)

	// Due tasks scheduler
	go func(tg *tg.TG) {
		fsBackend := afero.NewOsFs()
		var lastFrozenRequestCheckAt time.Time // We use this parameter to avoid logging the same frozen request many times
		for {
			select {
			case <-ticker.C:
				err := worker.MoveDueTasksToToday(config.Config.StoragePath, config.Config.ConfigFilename, fsBackend, telegram)
				if err != nil {
					fmt.Printf("Worker's error: %s\n", err)
				}
				reqs := userLocker.FrozenRequests(time.Second, lastFrozenRequestCheckAt.Add(-time.Second))
				lastFrozenRequestCheckAt = time.Now()
				for userID, req := range reqs {
					slog.Error("Frozen request", "userID", userID, "req", req)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}(telegram)

	go habitsServer()

	// Service
	tgConfig := tgbotapi.NewUpdate(0)
	tgConfig.Timeout = 60
	updates := api.GetUpdatesChan(tgConfig)

	for upd := range updates {
		go func(upd tgbotapi.Update) {
			defer func() {
				err := recover()
				if err != nil {
					slog.Error("Bot panic", "err", err)
				}
			}()

			u := tg.NewUpd(upd)
			userID := u.UserID()

			var updJSON []byte
			updJSON, _ = json.Marshal(upd)
			slog.Debug("Bot update: ", "upd", string(updJSON))

			userLocker.Lock(userID, string(updJSON))
			defer userLocker.Unlock(userID)

			userPath := path.Join(config.Config.StoragePath, txt.I64(userID))
			userFS, err := fs.NewFS(userPath, afero.NewOsFs())
			if err != nil {
				slog.Error("Bot error: can't create fs", "err", err)
				return
			}
			err = userFS.CreateDirsIfNotExist()
			if err != nil {
				slog.Error("Bot error: can't create user dirs", "err", err)
				return
			}

			userconfPath := userFS.UnsafePath("", config.Config.ConfigFilename)
			userconf := userconfig.NewConfig(userID, userconfPath)
			err = userconf.CreateDefaultIfNotExists()
			if err != nil {
				slog.Error("Bot error: can't create default user config", "err", err)
				return
			}

			bot := internal.NewBot(userID, telegram, userFS, db.NewDB(), userconf)
			if err := bot.Answer(u); err != nil {
				slog.Error("Bot error", "err", err)
			}
		}(upd)
	}
}
