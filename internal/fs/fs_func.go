package fs

import (
	"crypto/md5"
	"encoding/hex"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"zakirullin/stuffbot/pkg/txt"
)

func SanitizeFilename(filename string) string {
	// Under Linux and other Unix-related systems,
	// there are only two characters that cannot
	// appear in the name of a file or directory,
	// and those are NUL '\0' and slash '/'.
	// For Windows we only handle '\',
	// consider sanitazing other characters
	filename = strings.ReplaceAll(filename, "\x00", "")
	filename = strings.ReplaceAll(filename, "/", escapedForwardSlash)
	filename = strings.ReplaceAll(filename, "\\", escapedBackwardSlash)

	return filename
}

func UnsanitizeFilename(filename string) string {
	filename = strings.ReplaceAll(filename, escapedForwardSlash, "/")
	filename = strings.ReplaceAll(filename, escapedBackwardSlash, "\\")

	return filename
}

func Title(filename string) string {
	return txt.Ucfirst(strings.TrimSuffix(strings.TrimSpace(filename), FileExt))
}

func Hash(filename string) string {
	hash := md5.Sum([]byte(filename))
	return hex.EncodeToString(hash[:])[:11]
}

// ShortHash returns a short hash of the filename
// Telegram only allows 64 bytes in callback_data,
// so if we have 3 params we should use shortHash.
// More collisions are possible, but it's a trade-off.
func ShortHash(filename string) string {
	hash := md5.Sum([]byte(filename))
	return hex.EncodeToString(hash[:])[:5]
}

func ExcludeChecklists(dirs []File) []File {
	var newDirs []File
	for _, dir := range dirs {
		isChecklist := strings.HasPrefix(dir.Name, "_") && strings.HasSuffix(dir.Name, "_")
		if isChecklist {
			continue
		}

		newDirs = append(newDirs, dir)
	}

	return newDirs
}

func ExcludeSystemDirs(dirs []File) []File {
	var newDirs []File
	for _, dir := range dirs {
		if slices.Contains([]string{DirImg, DirArchive, DirJournal, DirInsights}, dir.Name) {
			continue
		}

		newDirs = append(newDirs, dir)
	}

	return newDirs
}

func ExcludeTaskDirs(dirs []File) []File {
	var newDirs []File
	for _, dir := range dirs {
		if slices.Contains([]string{DirToday, DirLater}, dir.Name) {
			continue
		}

		newDirs = append(newDirs, dir)
	}

	return newDirs
}

func ExcludePomodoro(files []File) []File {
	var newFiles []File
	for _, file := range files {
		if file.Name == FilePomodoro {
			continue
		}

		newFiles = append(newFiles, file)
	}

	return newFiles
}

func ExcludeConfig(files []File) []File {
	var newFiles []File
	for _, file := range files {
		if file.Name == FileConfig && file.ParentDir == DirRoot {
			continue
		}

		newFiles = append(newFiles, file)
	}

	return newFiles
}

func OnlyNoteDirs(dirs []File) []File {
	return ExcludeSystemDirs(ExcludeTaskDirs(ExcludeChecklists(dirs)))
}

func OnlyChecklists(dirs []File) []File {
	entries := OnlyDirs(ExcludeSystemDirs(ExcludeTaskDirs(dirs)))

	var dirsWithChecklists []File
	for _, entry := range entries {
		isChecklist := strings.HasPrefix(entry.Name, "_") && strings.HasSuffix(entry.Name, "_")
		if isChecklist {
			dirsWithChecklists = append(dirsWithChecklists, entry)
		}
	}

	return dirsWithChecklists
}

func OnlyMDFiles(entries []File) []File {
	var files []File
	for _, file := range entries {
		if file.IsDir {
			continue
		}

		if filepath.Ext(file.Name) != FileExt {
			continue
		}

		files = append(files, file)
	}

	return files
}

func OnlyDirs(entries []File) []File {
	var dirs []File
	for _, file := range entries {
		if !file.IsDir {
			continue
		}

		dirs = append(dirs, file)
	}

	return dirs
}

// OnlyUserDirs returns only directories that look like user IDs
func OnlyUserDirs(entries []File) []File {
	var dirs []File
	for _, file := range entries {
		if !file.IsDir {
			continue
		}
		if _, err := strconv.Atoi(file.Name); err != nil {
			continue
		}

		dirs = append(dirs, file)
	}

	return dirs
}

func OnlyFilenames(entries []File) []string {
	var filenames []string
	for _, entry := range entries {
		filenames = append(filenames, entry.Name)
	}

	return filenames
}

func SortByCtimeDesc(entries []File) []File {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Ctime > entries[j].Ctime
	})

	return entries
}
