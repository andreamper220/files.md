// Package fs provides a simple interface for files manipulations.
// Bot users should have all their artefacts saved in cross-platform
// plain text files, that's why we use good old-fashioned filesystem.
// Each user should have its own isolated root folder.
package fs

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/afero"
	"golang.org/x/exp/slices"

	"zakirullin/stuffbot/config"
	"zakirullin/stuffbot/pkg/txt"
)

var (
	errUnsafePath   = errors.New("unsafe path, possible security issue")
	errCannotUnhash = errors.New("cannot unhash, maybe the file is missing")
)

const (
	DirRoot     = ""
	DirArchive  = "archive"
	DirToday    = "today"
	DirLater    = "later"
	DirImg      = "img"
	DirJournal  = "journal"
	DirHabits   = "habits"
	DirInsights = "insights"
	DirRead     = "_read_"
	DirWatch    = "_watch_"
	DirShop     = "_shop_"

	FilePomodoro = "Finished a break.md"
	FileConfig   = "config.json"

	FileExt = ".md"

	minSearchSimilarity  = 70
	escapedForwardSlash  = "{|}"
	escapedBackwardSlash = "{||}"
)

// FS allows us to manipulate user files. We can use different
// backends, like an in-memory backend, which we use for testing.
// Check out types implementing afero.Fs for all available backends.
type FS struct {
	RootPath string
	backend  afero.Fs
}

// File represents a file or directory
type File struct {
	Name        string // Filename with extension
	Hash        string
	Title       string
	Ctime       int64
	IsMultiline bool
	IsDir       bool
	ParentDir   string
}

// NewUserFS creates a new FS for a specific user with os.FS backend.
func NewUserFS(userID int64) (*FS, error) {
	userAbsPath := path.Join(config.BotCfg.StorageDir, txt.I64(userID))
	backend := afero.NewOsFs()

	return NewFS(userAbsPath, backend)
}

func NewFS(absRootPath string, backend afero.Fs) (*FS, error) {
	exists, err := afero.Exists(backend, absRootPath)
	if err != nil {
		return nil, fmt.Errorf("new fs: %w", err)
	}
	if !exists {
		err = backend.Mkdir(absRootPath, 0o755)
		if err != nil {
			return nil, fmt.Errorf("new fs: %w", err)
		}
	}

	return &FS{absRootPath, backend}, nil
}

func NewFile(name, hash, title string, ctime int64, isMultiline, isDir bool, parentDir string) File {
	return File{name, hash, title, ctime, isMultiline, isDir, parentDir}
}

func (fs FS) CreateDirsIfNotExist() error {
	for _, dir := range []string{
		DirArchive,
		DirToday,
		DirLater,
		DirImg,
		DirRead,
		DirWatch,
		DirShop,
		DirHabits,
		DirJournal,
		DirInsights,
	} {
		userPath := path.Join(fs.RootPath, dir)
		exists, err := afero.Exists(fs.backend, userPath)
		if err != nil {
			return fmt.Errorf("create default dirs: %w", err)
		}

		if !exists {
			err = fs.backend.Mkdir(userPath, 0o755)
			if err != nil {
				return fmt.Errorf("create default dirs: %w", err)
			}
		}
	}

	return nil
}

func (fs FS) Exists(dir, filename string) (bool, error) {
	filePath := fs.UnsafePath(dir, filename)
	isSafe, err := fs.isSafe(filePath)
	if err != nil {
		return false, fmt.Errorf("exists: can't check if the file is safe to access '%s': %w", filePath, err)
	}
	if !isSafe {
		return false, fmt.Errorf("exists: unsafe path '%s': %w", filePath, errUnsafePath)
	}

	exists, err := afero.Exists(fs.backend, filePath)
	if err != nil {
		return false, fmt.Errorf("exists: can't check whether the file '%s'/'%s' exists: %w", dir, filename, err)
	}

	return exists, nil
}

func (fs FS) Read(dir, filename string) (string, error) {
	filePath := fs.UnsafePath(dir, filename)
	isSafe, err := fs.isSafe(filePath)
	if err != nil {
		return "", fmt.Errorf("fs read: can't check if the file is safe to access '%s': %w", filePath, err)
	}
	if !isSafe {
		return "", fmt.Errorf("fs read: unsafe filePath '%s': %w", filePath, errUnsafePath)
	}

	content, err := afero.ReadFile(fs.backend, filePath)
	if err != nil {
		return "", fmt.Errorf("fs read: can't read file '%s': %w", filePath, err)
	}

	return string(content), nil
}

func (fs FS) Write(dir, filename, content string) error {
	filePath := fs.UnsafePath(dir, filename)
	isSafe, err := fs.isSafe(filePath)
	if err != nil {
		return fmt.Errorf("fs write: check if file is safe to access '%s': %w", filePath, err)
	}

	if !isSafe {
		return fmt.Errorf("fs write: unsafe filePath '%s': %w", filePath, errUnsafePath)
	}

	dirs := strings.Split(filePath, "/")
	dirs = dirs[:len(dirs)-1]
	pathToDir := strings.Join(dirs, "/")
	if err := fs.backend.MkdirAll(pathToDir, 0o755); err != nil {
		return fmt.Errorf("fs write: can't create dirs '%s': %w", pathToDir, err)
	}

	// Append mode for forwards?
	if err := afero.WriteFile(fs.backend, filePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("fs write to '%s/%s': %w", dir, filename, err)
	}

	return nil
}

func (fs FS) MakeDir(dir string) error {
	filePath := fs.UnsafePath(dir, "")
	isSafe, err := fs.isSafe(filePath)
	if err != nil {
		return fmt.Errorf("fs make dir: check if file is safe to access '%s': %w", filePath, err)
	}
	if !isSafe {
		return fmt.Errorf("fs make dir: unsafe filePath '%s': %w", filePath, errUnsafePath)
	}

	err = fs.backend.Mkdir(filePath, 0o755)
	if err != nil {
		return fmt.Errorf("fs can't make dir: %w", err)
	}

	return nil
}

func (fs FS) Del(dir, filename string) error {
	filePath := fs.UnsafePath(dir, filename)
	isSafe, err := fs.isSafe(filePath)
	if err != nil {
		return fmt.Errorf("fs del: check if file is safe to access '%s': %w", filePath, err)
	}
	if !isSafe {
		return fmt.Errorf("fs del file: unsafe filePath '%s': %w", filePath, errUnsafePath)
	}

	err = fs.backend.Remove(filePath)
	if err != nil {
		return fmt.Errorf("fs file: can't remove '%s': %w", filePath, err)
	}

	return nil
}

// TODO check safety
func (fs FS) Rename(oldDir, oldFilename, newDir, newFilename string) error {
	oldPath := fs.UnsafePath(oldDir, oldFilename)
	isSafe, err := fs.isSafe(oldPath)
	if err != nil {
		return fmt.Errorf("fs rename: check if file is safe to access '%s': %w", oldPath, err)
	}
	if !isSafe {
		return fmt.Errorf("fs can't rename from '%s': %w", oldPath, errUnsafePath)
	}

	newPath := fs.UnsafePath(newDir, newFilename)
	isSafe, err = fs.isSafe(newPath)
	if err != nil {
		return fmt.Errorf("fs rename: check if file is safe to access '%s': %w", newPath, err)
	}
	if !isSafe {
		return fmt.Errorf("fs can't rename to '%s': %w", newPath, errUnsafePath)
	}

	err = fs.backend.Rename(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("can't rename from '%s' to '%s': %w", oldPath, newPath, err)
	}

	return nil
}

func (fs FS) Unhash(dir, filenameHash string) (string, error) {
	if dir == DirRoot && filenameHash == DirRoot {
		return DirRoot, nil
	}

	// TODO add safety checks (what safety checks?)

	filenames, err := fs.FilesAndDirs(dir)
	if err != nil {
		return "", fmt.Errorf("can't unhash: %w", err)
	}
	for _, file := range filenames {
		if strings.HasPrefix(fs.md5(file.Name), filenameHash) {
			return file.Name, nil
		}
	}

	// Fallback, treat hash as filename
	for _, file := range filenames {
		if strings.HasPrefix(file.Name, filenameHash) {
			return file.Name, nil
		}
	}

	return "", fmt.Errorf("can't unhash '%s' in '%s': %w", filenameHash, dir, errCannotUnhash)
}

func (fs FS) FilesAndDirs(dir string) ([]File, error) {
	userPath := fs.UnsafePath(dir, "")
	isSafe, err := fs.isSafe(userPath)
	if err != nil {
		return nil, fmt.Errorf("exists: check if file is safe to access '%s': %w", userPath, err)
	}
	if !isSafe {
		return nil, fmt.Errorf("can't get files for '%s': %w", path.Join(fs.RootPath, dir), errUnsafePath)
	}

	entries, err := afero.ReadDir(fs.backend, userPath)
	if err != nil {
		return nil, fmt.Errorf("can't get files for '%s': %w", path.Join(fs.RootPath, dir), err)
	}

	var files []File
	// TODO remove gitignore
	ignoredFiles := []string{".", "..", ".obsidian", ".gitignore", ".DS_Store", ".git"}
	for _, entry := range entries {
		if slices.Contains(ignoredFiles, entry.Name()) {
			continue
		}

		file := NewFile(
			entry.Name(),
			Hash(entry.Name()),
			Title(entry.Name()),
			Ctime(entry),
			entry.Size() > 0,
			entry.IsDir(),
			dir,
		)
		files = append(files, file)
	}

	return files, nil
}

func (fs FS) Dirs() ([]File, error) {
	files, err := fs.FilesAndDirs(DirRoot)
	if err != nil {
		return nil, fmt.Errorf("can't get dirs: %w", err)
	}

	var dirs []File
	for _, file := range files {
		isDir, err := afero.IsDir(fs.backend, fs.UnsafePath(DirRoot, file.Name))
		if err != nil {
			return nil, fmt.Errorf("can't get dirs: %w", err)
		}
		if !isDir {
			continue
		}

		dirs = append(dirs, file)
	}

	return dirs, nil
}

// TODO test all FS' public the methods for UnsafePath traversal
// TODO after you cover everything with the tests, we may remove this method
// because we build our own paths (???)
// TODO release remove error?
// isSafe doesn't eval symlinks, so an attacker can create a symlink to a file
// outside the RootPath. If we use filepath.EvalSymlinks to expand symlinks and
// check the real path for safety - we are still prone to TOCTOU (time-of-check to time-of-use)
// attacks due to the race condition. The only real way to prevent this is to disallow symlinks
// at the OS level. We can do this by mounting a folder with nosymfollow flag, see README.md.
func (fs FS) isSafe(path string) (bool, error) {
	path = filepath.Clean(path)

	// Path traversal attack (filepath.Clean only cleans absolute paths from ../)
	// https://owasp.org/www-community/attacks/Path_Traversal
	// A better way would be to convert the path to absolute path, but AferoFS doesn't support that
	if strings.Contains(path, "../") || strings.Contains(path, "/..") {
		return false, nil
	}

	outsideOfUserRoot := !strings.HasPrefix(path, fs.RootPath)
	if outsideOfUserRoot {
		return false, nil
	}

	return true, nil
}

func (fs FS) IsMultiline(dir, filename string) (bool, error) {
	content, err := fs.Read(dir, filename)
	if err != nil {
		return false, fmt.Errorf("can't check for multiline: %w", err)
	}
	content = strings.TrimSpace(content)

	return len(content) > 0, nil
}

func (fs FS) md5(filename string) string {
	hash := md5.Sum([]byte(filename))
	return hex.EncodeToString(hash[:])[:11]
}

func Filename(title string) string {
	return txt.Ucfirst(title) + FileExt
}

func IsChecklistItem(filename string) bool {
	validChecklistItem := regexp.MustCompile(`^-.*?-(.+)`)

	return validChecklistItem.MatchString(filename)
}

// SearchFiles performs search among all user files
// Allowed query formats:
// "directory" - return all notes from directories prefixed by this directory
// "directory note_name" - search for this note_name in all matching directories
// "note_name" - search for this note_name across all directories
// "" - return all the notes
func (fs FS) SearchFiles(query string) ([]File, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	// Check for directory traversal attack
	if strings.Contains(query, "/") {
		return nil, fmt.Errorf("search notes: unsafe query '%s': %w", query, errUnsafePath)
	}

	var supposedDir, search string
	dirExists, err := fs.Exists(DirRoot, query)
	if err != nil {
		return nil, fmt.Errorf("search notes: %w", err)
	}
	if dirExists {
		supposedDir = query
	} else {
		parts := strings.SplitN(query, " ", 2)
		supposedDir = parts[0]
		if len(parts) > 1 {
			search = strings.TrimSpace(parts[1])
		}
	}

	// Find match by notes directory name
	var searchInDirs []string
	notesDirs, err := fs.FilesAndDirs(DirRoot)
	if err != nil {
		return nil, fmt.Errorf("search notes: %w", err)
	}
	notesDirs = OnlyNoteDirs(notesDirs)
	notesDirs = append(notesDirs, NewFile(DirRoot, "", DirRoot, 0, false, true, ""))
	notesDirs = append(notesDirs, NewFile(DirJournal, Hash(DirJournal), DirJournal, 0, false, true, DirRoot))
	notesDirs = append(notesDirs, NewFile(DirInsights, Hash(DirInsights), DirInsights, 0, false, true, DirRoot))
	for _, noteDir := range notesDirs {
		if strings.HasPrefix(noteDir.Name, supposedDir) {
			searchInDirs = append(searchInDirs, noteDir.Name)
		}
	}

	// If no matching directories are found, we search through all directories
	if len(searchInDirs) == 0 {
		for _, noteDir := range notesDirs {
			searchInDirs = append(searchInDirs, noteDir.Name)
		}
		search = query
	}

	var notes []File
	for _, dir := range searchInDirs {
		// We can tolerate incomplete search
		files, _ := fs.FilesAndDirs(dir)
		files = OnlyMDFiles(files)
		notes = append(notes, files...)
	}
	notes = SortByCtimeDesc(notes)

	var matchedNotes []File
	for _, note := range notes {
		isWildcard := len(search) == 0
		isSubstring := strings.Contains(strings.ToLower(note.Title), search)
		isSimilar := txt.Similar(strings.ToLower(note.Title), search) > minSearchSimilarity
		if isWildcard || isSubstring || isSimilar {
			matchedNotes = append(matchedNotes, note)
		}
	}

	return matchedNotes, nil
}

// TODO check if safe
// Touch updates an existing file's access and modification times.
// If there's no such file it creates an empty file.
func (fs FS) Touch(dir, filename string) error {
	exists, err := fs.Exists(dir, filename)
	if err != nil {
		return fmt.Errorf("touch: %w", err)
	}
	if exists {
		err = fs.backend.Chtimes(fs.UnsafePath(dir, filename), time.Now(), time.Now())
		if err != nil {
			return fmt.Errorf("touch: can't update file's ctime: %w", err)
		}
		return nil
	}
	err = fs.Write(dir, filename, "")
	if err != nil {
		return fmt.Errorf("touch: can't create empty file: %w", err)
	}
	return nil
}

// UnsafePath builds a user-specific path.
// It'S NOT SAFE to use this method with user input.
// Sanitize Early, call SanitizeFilename
// as soon as you get on dir and filename from user input
func (fs FS) UnsafePath(dir, filename string) string {
	p := path.Join(fs.RootPath, dir, filename)

	return p
}

func (fs FS) Ctime(dir, filename string) (int64, error) {
	filePath := fs.UnsafePath(dir, filename)
	isSafe, err := fs.isSafe(filePath)
	if err != nil {
		return 0, fmt.Errorf("fs file: can't check if the file is safe to access '%s': %w", filePath, err)
	}
	if !isSafe {
		return 0, fmt.Errorf("fs file: unsafe filePath '%s': %w", filePath, errUnsafePath)
	}

	info, err := fs.backend.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("fs file: can't stat file '%s': %w", filePath, err)
	}

	return Ctime(info), nil
}

// Ctimes recursively scans a directory and returns the ctime
// for all .md files as Unix timestamps.
func (fs FS) Ctimes() (map[string]int64, error) {
	timestamps := make(map[string]int64)

	err := afero.Walk(fs.backend, fs.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") && path != fs.RootPath {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(strings.ToLower(path), FileExt) {
			return nil
		}

		relPath, err := filepath.Rel(fs.RootPath, path)
		if err != nil {
			return nil
		}

		if relPath == "" {
			relPath = "."
		}

		timestamps[relPath] = Ctime(info)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return timestamps, nil
}
