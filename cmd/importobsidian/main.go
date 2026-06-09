// One-time import from an Obsidian vault into a Files.md storage directory.
//
// Usage:
//
//	go run ./cmd/importobsidian --src "C:\Users\ADMIN\Documents\obsidian" --dst ./storage/USER_ID
//
// Daily notes from Daily Notes/Быстрые заметки/YYYYMMDD.md are merged into
// journal/YYYY.MM Month.md files used by Files.md.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	dailyNoteRE = regexp.MustCompile(`^(\d{4})(\d{2})(\d{2})\.md$`)
	wikiRE      = regexp.MustCompile(`\[\[([^\[\]]+)\]\]`)
)

func main() {
	src := flag.String("src", `C:\Users\ADMIN\Documents\obsidian`, "Obsidian vault path")
	dst := flag.String("dst", "", "Files.md user storage directory (required)")
	dryRun := flag.Bool("dry-run", false, "Print actions without writing")
	flag.Parse()

	if *dst == "" {
		fmt.Fprintln(os.Stderr, "--dst is required")
		os.Exit(1)
	}

	if err := run(*src, *dst, *dryRun); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(src, dst string, dryRun bool) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	if err := os.MkdirAll(filepath.Join(dst, "journal"), 0o755); err != nil {
		return err
	}
	for _, dir := range []string{"media", "archive", "habits", "insights"} {
		if err := os.MkdirAll(filepath.Join(dst, dir), 0o755); err != nil {
			return err
		}
	}

	dailyDir := filepath.Join(src, "Daily Notes", "Быстрые заметки")
	journalFiles := map[string]string{}

	err := filepath.Walk(dailyDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		m := dailyNoteRE.FindStringSubmatch(info.Name())
		if m == nil {
			return nil
		}
		year, month, day := m[1], m[2], m[3]
		date, err := time.Parse("2006-01-02", fmt.Sprintf("%s-%s-%s", year, month, day))
		if err != nil {
			return nil
		}

		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		entry := strings.TrimSpace(string(body))
		if entry == "" {
			return nil
		}
		entry = convertWikilinks(entry, src)

		monthFile := date.Format("2006.01 January") + ".md"
		dayHeader := fmt.Sprintf("## %d %s, %s", date.Day(), date.Format("January"), date.Weekday())
		block := dayHeader + "\n" + entry + "\n"
		journalFiles[monthFile] = mergeJournalBlock(journalFiles[monthFile], dayHeader, block)
		return nil
	})
	if err != nil {
		return fmt.Errorf("import daily notes: %w", err)
	}

	for monthFile, content := range journalFiles {
		target := filepath.Join(dst, "journal", monthFile)
		if dryRun {
			fmt.Printf("journal: %s (%d bytes)\n", target, len(content))
			continue
		}
		if err := os.WriteFile(target, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
			return err
		}
		fmt.Printf("wrote %s\n", target)
	}

	skipDirs := map[string]bool{
		".obsidian":   true,
		".git":        true,
		"Templates":   true,
		"Daily Notes": true,
	}

	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		parts := strings.Split(rel, string(os.PathSeparator))
		if len(parts) > 0 && skipDirs[parts[0]] {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			targetDir := filepath.Join(dst, filepath.ToSlash(rel))
			if dryRun {
				fmt.Printf("mkdir: %s\n", targetDir)
				return nil
			}
			return os.MkdirAll(targetDir, 0o755)
		}

		if !strings.EqualFold(filepath.Ext(path), ".md") {
			return copyAsset(path, dst, dryRun)
		}

		if isSkippedRootFile(rel) {
			return nil
		}

		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		converted := convertWikilinks(string(body), src)
		target := filepath.Join(dst, filepath.ToSlash(rel))
		if dryRun {
			fmt.Printf("note: %s\n", target)
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(converted), 0o644); err != nil {
			return err
		}
		fmt.Printf("wrote %s\n", target)
		return nil
	})
	if err != nil {
		return fmt.Errorf("import notes: %w", err)
	}

	fmt.Println("import complete")
	return nil
}

func mergeJournalBlock(existing, dayHeader, block string) string {
	if existing == "" {
		return block
	}
	if strings.Contains(existing, dayHeader) {
		return existing + "\n" + strings.TrimPrefix(block, dayHeader+"\n")
	}
	return strings.TrimSpace(existing) + "\n\n" + block
}

func isSkippedRootFile(rel string) bool {
	switch strings.ToLower(filepath.Base(rel)) {
	case "welcome.md", "__new.md", "__recent.md":
		return true
	}
	return false
}

func copyAsset(srcPath, dstRoot string, dryRun bool) error {
	ext := strings.ToLower(filepath.Ext(srcPath))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".mp3", ".wav", ".ogg", ".mp4", ".pdf":
	default:
		return nil
	}

	dstPath := filepath.Join(dstRoot, "media", filepath.Base(srcPath))

	if dryRun {
		fmt.Printf("media: %s -> %s\n", srcPath, dstPath)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	return os.WriteFile(dstPath, data, 0o644)
}

func convertWikilinks(content, vaultRoot string) string {
	candidates := map[string]string{}
	filepath.Walk(vaultRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.EqualFold(filepath.Ext(path), ".md") {
			return nil
		}
		rel, _ := filepath.Rel(vaultRoot, path)
		name := strings.TrimSuffix(filepath.Base(rel), ".md")
		candidates[name] = filepath.ToSlash(rel)
		return nil
	})

	return wikiRE.ReplaceAllStringFunc(content, func(match string) string {
		name := match[2 : len(match)-2]
		if idx := strings.Index(name, "|"); idx != -1 {
			name = name[:idx]
		}
		target, ok := candidates[strings.TrimSpace(name)]
		if !ok {
			return match
		}
		label := strings.TrimSuffix(filepath.Base(target), ".md")
		return fmt.Sprintf("[%s](%s)", label, target)
	})
}
