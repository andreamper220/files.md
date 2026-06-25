package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadDotEnv loads variables from a .env file in the project root.
// It checks the working directory, the executable directory, and their parents.
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
	fmt.Println("No .env file found, relying on process environment")
}
