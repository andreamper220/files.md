package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/alicebob/miniredis/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
	"github.com/spf13/afero"
	"golang.org/x/exp/slog"

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
	conf, err := internal.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Error loading conf: %s\n", err))
	}

	// TODO move to config
	err = i18n.LoadLangFile("i18n/ru.json")
	if err != nil {
		panic(fmt.Sprintf("Error loading i18n: %s\n", err))
	}
	// TODO move to config
	err = i18n.LoadEmojiFile("i18n/emojis.json")
	if err != nil {
		panic(fmt.Sprintf("Error loading emoji: %s\n", err))
	}

	api, err := tgbotapi.NewBotAPI(conf.BotAPIToken)
	if err != nil {
		panic(fmt.Sprintf("Can't create TG api: %s\n", err))
	}
	telegram := tg.NewTG(api)

	redis, err := miniredis.Run()
	if err != nil {
		panic(fmt.Sprintf("Can't create Redis: %s\n", err))
	}
	defer redis.Close()

	// Initiate per-user locker
	userLocker := sync.NewPerUserLocker()

	// Workers
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	defer func(quit chan struct{}) {
		close(quit)
	}(quit)

	// Due tasks scheduler
	go func(redis *miniredis.Miniredis, tg *tg.TG) {
		fsBackend := afero.NewOsFs()
		var lastFrozenRequestCheckAt time.Time // We use this parameter to avoid logging the same frozen request many times
		for {
			select {
			case <-ticker.C:
				err := worker.MoveDueTasksToToday(conf.StoragePath, conf.ConfigFilename, fsBackend)
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
	}(redis, telegram)

	// TODO graceful shutdown etc
	go func() {
		router := http.NewServeMux()
		router.HandleFunc("GET /habits/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("id")
			fmt.Fprintf(w, "habits task with id=%v\n", id)
		})

		http.ListenAndServe(":8080", router)
	}()

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

			userPath := path.Join(conf.StoragePath, txt.I64(userID))
			userFS, err := fs.NewFS(userPath, afero.NewOsFs())
			if err != nil {
				slog.Error("Bot error: can't create fs", "err", err)
				return
			}
			err = userFS.CreateUserDirs()
			if err != nil {
				slog.Error("Bot error: can't create user dirs", "err", err)
				return
			}

			userconf := userconfig.NewConfig()
			userconfPath := userFS.Path("", conf.ConfigFilename)
			err = userconf.LoadOrCreate(userconfPath)
			if err != nil {
				slog.Error("Bot error: can't get or create conf", "err", err)
				return
			}
			defer func() {
				err = userconf.Save(userconfPath)
				if err != nil {
					slog.Error("Bot error: can't save userconfig", "err", err)
				}
			}()

			bot := internal.NewBot(userID, telegram, userFS, db.NewDB(redis), userconf)

			if err := bot.Reply(u); err != nil {
				slog.Error("Bot error", "err", err)
			}
		}(upd)
	}
}
