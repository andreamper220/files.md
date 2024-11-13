// Provides a server for habits tracking functionality through Telegram miniapps.
// SSLs certificates are handled automatically via LetsEncrypt.
package main

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"
	"golang.org/x/crypto/acme/autocert"

	"zakirullin/stuffbot/config"
	"zakirullin/stuffbot/i18n"
	"zakirullin/stuffbot/internal/fs"
	"zakirullin/stuffbot/internal/habits"
	"zakirullin/stuffbot/internal/journal"
	"zakirullin/stuffbot/internal/userconfig"
	"zakirullin/stuffbot/pkg/txt"
)

//go:embed templates/index.html
var html string

//go:embed templates/favicon.ico
var favicon string

//go:embed templates/styles.css
var styles string

//go:embed templates/tomassanchez.webp
var img string

// TODO release graceful shutdown etc
func habitsServer(habitsHost, certDir, logFilename string) {
	autocertManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(habitsHost, "www.files.md", "app.files.md"),
		Cache:      autocert.DirCache(certDir),
	}

	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatalf("Server: failed to open log file: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "Server Error: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Listen for HTTP requests on port 80 in a new goroutine. Use
	// autocertManager.HTTPHandler(nil) as the handler. This will send ACME
	// "http-01" challenge responses as necessary, and 302 redirect all other
	// requests to HTTPS.
	go func() {
		srv := &http.Server{
			Addr:         ":80",
			Handler:      autocertManager.HTTPHandler(nil),
			IdleTimeout:  time.Minute,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			ErrorLog:     logger,
		}

		err = srv.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()

	// Configure the TLS config to use the autocertManager.GetCertificate function.
	tlsConfig := &tls.Config{
		GetCertificate:   autocertManager.GetCertificate,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	router := http.NewServeMux()
	setupRouter(router, logger)
	srv := &http.Server{
		Addr:         ":443",
		Handler:      router,
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     logger,
	}

	err = srv.ListenAndServeTLS("", "") // Key and cert provided automatically by autocert
	if err != nil {
		panic(err)
	}
}

func setupRouter(router *http.ServeMux, logger *log.Logger) {
	router.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./editor/editor.html")
	})

	router.HandleFunc("/app/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/app/" {
			http.ServeFile(w, r, "./editor/editor.html")
			return
		}

		http.StripPrefix("/app/", http.FileServer(http.Dir("./editor"))).ServeHTTP(w, r)
	})

	//router.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
	//	if r.URL.Path == "/app" {
	//		http.ServeFile(w, r, "./editor/editor.html")
	//		return
	//	}
	//
	//	http.StripPrefix("/app", http.FileServer(http.Dir("./editor"))).ServeHTTP(w, r)
	//})
	//router.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
	//	http.ServeFile(w, r, "./editor/editor.html")
	//})

	// TODO add hashing or secrets
	// TODO before release habits_v2 => habits
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serving editor app
		host := r.Host
		if strings.HasPrefix(host, "app.") {
			if r.URL.Path == "" || r.URL.Path == "/" {
				http.ServeFile(w, r, "./app/app.html")
				return
			}

			http.FileServer(http.Dir("./app")).ServeHTTP(w, r)
			return
		}

		// Serving the site
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(html))
		if err != nil {
			logger.Printf("failed to write site response: %v", err)
		}
	})

	router.HandleFunc("/article", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://minds.md/zakirullin/cognitive", http.StatusFound)
	})

	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(favicon))
	})

	router.HandleFunc("/tomassanchez.webp", func(w http.ResponseWriter, r *http.Request) {
		oneMonth := 30 * 24 * time.Hour
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(oneMonth.Seconds())))
		w.Header().Set("Content-Type", "image/webp")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(img))
	})

	router.HandleFunc("/styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(styles))
	})

	router.HandleFunc("GET /habits_v2/{userID}", func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(r.PathValue("userID"), 10, 64)
		if err != nil {
			logger.Printf("failed to parse userID for habits: %v", err)
			_, _ = w.Write([]byte("can't parse userID"))
		}

		userPath := path.Join(config.BotCfg.StorageDir, txt.I64(userID))
		userFS, err := fs.NewFS(userPath, afero.NewOsFs())
		if err != nil {
			logger.Printf("failed to init userFS: %v", err)
			_, _ = w.Write([]byte("can't init userFS"))
		}

		str, err := habits.Render(userID, userFS)
		if err != nil {
			logger.Printf("failed to render habits: %v", err)
			_, _ = w.Write([]byte(err.Error()))
		}
		_, err = w.Write(str)
		if err != nil {
			logger.Printf("failed to write habits response: %v", err)
		}
	})

	router.HandleFunc("POST /habits_v2/{userID}/{habitName}/{yearDay}/{status}", func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(r.PathValue("userID"), 10, 64)
		if err != nil {
			logger.Printf("failed to parse userID: %v", err)
			_, _ = w.Write([]byte("can't parse userID"))
		}

		yearDay, err := strconv.ParseInt(r.PathValue("yearDay"), 10, 32)
		if err != nil {
			logger.Printf("failed to parse yearDay: %v", err)
			_, _ = w.Write([]byte("can't parse yearDay"))
		}

		status, err := strconv.ParseInt(r.PathValue("status"), 10, 32)
		if err != nil {
			logger.Printf("failed to parse status: %v", err)
			_, _ = w.Write([]byte("can't parse status"))
		}

		habitName := r.PathValue("habitName")

		userPath := path.Join(config.BotCfg.StorageDir, txt.I64(userID))
		userFS, err := fs.NewFS(userPath, afero.NewOsFs())
		if err != nil {
			logger.Printf("failed to init user fs: %v", err)
			_, _ = w.Write([]byte("can't init user fs"))
		}

		userHabits, err := habits.Habits(userFS, time.Now().Year())
		if err != nil {
			logger.Printf("failed to read habits: %v", err)
			_, _ = w.Write([]byte("can't read habits"))
		}

		if _, ok := userHabits[habitName]; !ok {
			userHabits[habitName] = make(habits.Year)
		}
		userHabits[habitName][int(yearDay)] = int(status)
		err = habits.Write(userFS, time.Now().Year(), userHabits)
		if err != nil {
			logger.Printf("failed to write habits: %v", err)
			_, _ = w.Write([]byte("can't write habits"))
		}

		var emoji string
		if habitName == habits.MoodHabit {
			if int(status) < len(habits.MoodEmojis) {
				emoji = habits.MoodEmojis[status]
			}
		} else {
			emoji, _ = userFS.Read(fs.DirHabits, fs.Filename(habitName))
			if emoji == "" {
				emoji = i18n.Emoji(habitName)
			}
			if emoji == "" {
				emoji = "⚡️"
			}
		}

		userConf := userconfig.NewConfig(userFS, userID, config.BotCfg.ConfigFilename)
		err = journal.AddEmoji(userFS, emoji, userConf.Timezone())
		if err != nil {
			logger.Printf("failed to write habit emoji to journal: %v", err)
			_, _ = w.Write([]byte("can't write habit emoji to journal"))
		}

		record := fmt.Sprintf("%s %s", emoji, habitName)
		err = journal.AddRecord(userFS, record, userConf.Timezone())
		if err != nil {
			logger.Printf("failed to write habit to journal: %v", err)
			_, _ = w.Write([]byte("can't write habit to journal"))
		}
	})
}
