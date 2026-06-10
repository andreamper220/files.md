package fs

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/afero"

	"github.com/zakirullin/files.md/server/config"
)

// NoteFile is a user-authored markdown note with filesystem metadata.
type NoteFile struct {
	ParentDir string
	Name      string
	Ctime     int64
	Mtime     int64
}

// DisplayPath returns a human-readable note path without the .md extension.
func (n NoteFile) DisplayPath() string {
	name := DisplayName(n.Name)
	parent := strings.Trim(n.ParentDir, "/")
	if parent == "" || parent == DirUserRoot {
		return name
	}
	if parent == DirJournal {
		return "журнал/" + name
	}
	return parent + "/" + name
}

// AllNoteFiles returns every user note in the vault (recursive, excluding system files).
func (fs FS) AllNoteFiles() ([]NoteFile, error) {
	rootPath, err := fs.SafePath(DirUserRoot, "")
	if err != nil {
		return nil, err
	}

	var notes []NoteFile
	err = afero.Walk(fs.backend, rootPath, func(p string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		base := filepath.Base(p)
		if strings.HasPrefix(base, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			rel, _ := filepath.Rel(rootPath, p)
			if rel != "." && isSkippedNoteDir(base) {
				return filepath.SkipDir
			}
			if rel != "." && strings.Count(rel, string(filepath.Separator)) >= 10 {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(base) != MDExt {
			return nil
		}

		relativeDir, _ := filepath.Rel(rootPath, filepath.Dir(p))
		if relativeDir == "." {
			relativeDir = DirUserRoot
		}
		relativeDir = filepath.ToSlash(relativeDir)

		if isSystemNoteFile(base, relativeDir) {
			return nil
		}

		notes = append(notes, NoteFile{
			ParentDir: relativeDir,
			Name:      base,
			Ctime:     Ctime(info),
			Mtime:     Mtime(info),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return notes, nil
}

func isSkippedNoteDir(name string) bool {
	return strings.EqualFold(name, DirMedia) ||
		strings.EqualFold(name, DirArchive) ||
		strings.EqualFold(name, DirHabits) ||
		strings.EqualFold(name, DirInsights)
}

func isSystemNoteFile(name, parentDir string) bool {
	systemFiles := []string{
		ChatFilename,
		LaterFilename,
		DoneFilename,
		ShopFilename,
		WatchFilename,
		ReadFilename,
		config.ServerCfg.ConfigFilename,
	}
	if slices.Contains(systemFiles, name) {
		return true
	}
	base := strings.TrimSuffix(name, MDExt)
	if strings.HasSuffix(base, "_") {
		return true
	}
	return false
}
