package fs

import (
	"fmt"
	"strings"
)

func isSystemDirName(name string) bool {
	return strings.EqualFold(name, DirMedia) ||
		strings.EqualFold(name, DirArchive) ||
		strings.EqualFold(name, DirJournal) ||
		strings.EqualFold(name, DirHabits) ||
		strings.EqualFold(name, DirInsights)
}

// JoinDir joins a parent directory path with a child name.
func JoinDir(parent, name string) string {
	parent = strings.Trim(parent, "/")
	if parent == "" || parent == DirUserRoot {
		return name
	}
	return parent + "/" + name
}

// ParentDirPath returns the parent of a nested directory path.
func ParentDirPath(dir string) string {
	dir = strings.Trim(dir, "/")
	if dir == "" || !strings.Contains(dir, "/") {
		return ""
	}
	return dir[:strings.LastIndex(dir, "/")]
}

// AllNoteDirPaths returns all user note directories recursively.
func (fs FS) AllNoteDirPaths() ([]string, error) {
	var result []string
	var walk func(parent string) error
	walk = func(parent string) error {
		entries, err := fs.FilesAndDirs(parent)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if !entry.IsDir || isSystemDirName(entry.Name) {
				continue
			}
			fullPath := JoinDir(parent, entry.Name)
			if parent == DirUserRoot {
				fullPath = entry.Name
			}
			result = append(result, fullPath)
			if err := walk(fullPath); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(DirUserRoot); err != nil {
		return nil, err
	}
	return result, nil
}

// FindNoteDirByShortHash resolves a directory path from a short hash or literal path.
func (fs FS) FindNoteDirByShortHash(hash string) (string, error) {
	if hash == "" {
		return DirUserRoot, nil
	}
	if strings.Contains(hash, "/") {
		return hash, nil
	}

	if name, err := fs.Unhash(DirUserRoot, hash); err == nil {
		isDir, err := fs.Exists(name, "")
		if err == nil && isDir {
			return name, nil
		}
	}

	paths, err := fs.AllNoteDirPaths()
	if err != nil {
		return "", err
	}
	for _, dirPath := range paths {
		if ShortHash(dirPath) == hash || strings.HasPrefix(Hash(dirPath), hash) {
			return dirPath, nil
		}
	}

	return "", fmt.Errorf("can't find dir for hash '%s': %w", hash, ErrCannotUnhash)
}

// ResolveDirParam converts a callback dir parameter into a filesystem directory path.
func (fs FS) ResolveDirParam(dirParam string) (string, error) {
	if dirParam == "" || dirParam == DirUserRoot || dirParam == "/" {
		return DirUserRoot, nil
	}
	if strings.Contains(dirParam, "/") {
		return dirParam, nil
	}
	return fs.FindNoteDirByShortHash(dirParam)
}
