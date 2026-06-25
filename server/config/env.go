package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadDotEnv loads variables from a .env file when present.
// Under Docker Compose, variables are usually injected via env_file; a missing
// file on disk is normal as long as the process environment is populated.
func LoadDotEnv() {
	seen := map[string]bool{}
	var candidates []string
	add := func(p string) {
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		candidates = append(candidates, p)
	}

	add(".env")
	add("/app/.env")
	if wd, err := os.Getwd(); err == nil {
		add(filepath.Join(wd, ".env"))
		add(filepath.Join(wd, "..", ".env"))
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		add(filepath.Join(dir, ".env"))
		add(filepath.Join(dir, "..", ".env"))
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if err := godotenv.Load(path); err == nil {
			fmt.Println("Loaded .env from", path)
			return
		}
	}
	if envLooksConfigured() {
		fmt.Println("Using environment variables (no .env file on disk)")
		return
	}
	fmt.Println("No .env file found and BOT_API_TOKEN / KIE_API_KEY are not set")
}

func envLooksConfigured() bool {
	return os.Getenv("BOT_API_TOKEN") != "" || os.Getenv("KIE_API_KEY") != ""
}

// LogStartupConfig logs whether required settings are present (never logs secrets).
func LogStartupConfig() {
	if ServerCfg.BotAPIToken != "" {
		slog.Info("config ready", "telegram_bot", true)
	} else {
		slog.Warn("config missing", "telegram_bot", false, "hint", "set BOT_API_TOKEN in .env or environment")
	}
	if ServerCfg.KieAPIKey != "" {
		slog.Info("config ready", "voice_transcription", true)
	} else {
		slog.Warn("config missing", "voice_transcription", false, "hint", "set KIE_API_KEY in .env or environment")
	}
	if ServerCfg.TokensSalt != "" {
		slog.Info("config ready", "tokens_salt", true)
	} else {
		slog.Warn("config missing", "tokens_salt", false, "hint", "set TOKENS_SALT in .env or environment")
	}
}
