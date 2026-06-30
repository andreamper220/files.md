package life

import (
	"fmt"
	"strings"

	"github.com/zakirullin/files.md/server/fs"
)

// MigrateOptions configures a one-time import from legacy note folders into life areas.
type MigrateOptions struct {
	Sphere      string
	Kind        Kind
	Dirs        []string // empty = all legacy note dirs at user root
	RootMD      bool
	RootProject string
	Copy        bool
	DryRun      bool
}

// MigrateResult summarizes a migration run.
type MigrateResult struct {
	ProjectsCreated int
	SectionsCreated int
	FilesMoved      int
	FilesSkipped    int
	DirsProcessed   int
}

// MigrateFromFolders moves legacy category folders (notes/, brain/, …) into spheres/…/areas.
// Each top-level folder becomes a project; nested folders become sections; notes land in the chosen kind subdir.
func MigrateFromFolders(fsys *fs.FS, opts MigrateOptions) (MigrateResult, error) {
	var res MigrateResult
	if opts.Sphere == "" {
		opts.Sphere = "Личное"
	}
	if opts.RootProject == "" {
		opts.RootProject = "Inbox"
	}
	if opts.Kind != KindDraft && opts.Kind != KindFinal && opts.Kind != KindDiscussion {
		opts.Kind = KindFinal
	}

	if err := Init(fsys); err != nil {
		return res, fmt.Errorf("migrate: init life: %w", err)
	}

	spherePath := SpherePath(opts.Sphere)
	exists, err := fsys.Exists(spherePath, "")
	if err != nil {
		return res, err
	}
	if !exists {
		return res, fmt.Errorf("migrate: sphere %q does not exist; run life init first", opts.Sphere)
	}

	targetDirs, err := legacyNoteDirs(fsys, opts.Dirs)
	if err != nil {
		return res, err
	}

	for _, dirName := range targetDirs {
		res.DirsProcessed++
		oldDir := dirName
		projectPath, created, err := ensureProject(fsys, spherePath, dirName, opts.DryRun)
		if err != nil {
			return res, fmt.Errorf("migrate %s: %w", dirName, err)
		}
		if created {
			res.ProjectsCreated++
		}
		if err := migrateTree(fsys, oldDir, projectPath, opts.Kind, opts.Copy, opts.DryRun, &res); err != nil {
			return res, fmt.Errorf("migrate %s: %w", dirName, err)
		}
		if !opts.Copy && !opts.DryRun {
			if err := fsys.DelDir(oldDir); err != nil {
				return res, fmt.Errorf("migrate %s: can't remove source dir: %w", dirName, err)
			}
		}
	}

	if opts.RootMD {
		projectPath, created, err := ensureProject(fsys, spherePath, opts.RootProject, opts.DryRun)
		if err != nil {
			return res, fmt.Errorf("migrate root notes: %w", err)
		}
		if created {
			res.ProjectsCreated++
		}
		entries, err := fsys.FilesAndDirs(fs.DirUserRoot)
		if err != nil {
			return res, err
		}
		for _, file := range fs.OnlyFiles(entries) {
			if !strings.HasSuffix(file.Name, fs.MDExt) {
				continue
			}
			if isLegacySystemFile(file.Name) {
				continue
			}
			if err := migrateFile(fsys, fs.DirUserRoot, file.Name, projectPath, opts.Kind, opts.Copy, opts.DryRun, &res); err != nil {
				return res, fmt.Errorf("migrate root %s: %w", file.Name, err)
			}
		}
	}

	return res, nil
}

func legacyNoteDirs(fsys *fs.FS, selected []string) ([]string, error) {
	entries, err := fsys.FilesAndDirs(fs.DirUserRoot)
	if err != nil {
		return nil, err
	}
	all := fs.OnlyNoteDirs(fs.OnlyDirs(entries))
	var names []string
	for _, dir := range all {
		if isLegacySkipDir(dir.Name) {
			continue
		}
		names = append(names, dir.Name)
	}
	if len(selected) == 0 {
		return names, nil
	}
	want := make(map[string]bool, len(selected))
	for _, name := range selected {
		want[strings.TrimSpace(name)] = true
	}
	var filtered []string
	for _, name := range names {
		if want[name] {
			filtered = append(filtered, name)
		}
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("migrate: none of the requested dirs found among legacy note folders")
	}
	return filtered, nil
}

// LegacyNoteDirNames lists legacy note folder names at the user root.
func LegacyNoteDirNames(fsys *fs.FS, selected []string) ([]string, error) {
	return legacyNoteDirs(fsys, selected)
}

func isLegacySkipDir(name string) bool {
	switch name {
	case DirSpheres, fs.DirHabits, "img", ".obsidian", ".git":
		return true
	default:
		return false
	}
}

func isLegacySystemFile(name string) bool {
	switch name {
	case fs.ChatFilename, fs.LaterFilename, fs.DoneFilename,
		fs.ShopFilename, fs.WatchFilename, fs.ReadFilename,
		IndexFilename, SphereHubFile, ProjectHubFile, TasksFilename:
		return true
	default:
		return strings.HasSuffix(strings.TrimSuffix(name, fs.MDExt), "_")
	}
}

func ensureProject(fsys *fs.FS, spherePath, name string, dryRun bool) (string, bool, error) {
	projectPath := ProjectPath(spherePath, name)
	exists, err := fsys.Exists(projectPath, "")
	if err != nil {
		return "", false, err
	}
	if exists {
		if err := EnsureAreaDirs(fsys, projectPath); err != nil {
			return "", false, err
		}
		return projectPath, false, nil
	}
	if dryRun {
		return projectPath, true, nil
	}
	createdPath, err := CreateProject(fsys, spherePath, name)
	return createdPath, true, err
}

func ensureSection(fsys *fs.FS, parentPath, name string, dryRun bool) (string, bool, error) {
	sectionPath := fs.JoinDir(parentPath, fs.SanitizeFilename(name))
	exists, err := fsys.Exists(sectionPath, "")
	if err != nil {
		return "", false, err
	}
	if exists {
		if err := EnsureAreaDirs(fsys, sectionPath); err != nil {
			return "", false, err
		}
		return sectionPath, false, nil
	}
	if dryRun {
		return sectionPath, true, nil
	}
	createdPath, err := CreateSection(fsys, parentPath, name)
	return createdPath, true, err
}

func migrateTree(fsys *fs.FS, srcDir, areaPath string, kind Kind, copy, dryRun bool, res *MigrateResult) error {
	entries, err := fsys.FilesAndDirs(srcDir)
	if err != nil {
		return err
	}

	for _, file := range fs.OnlyFiles(entries) {
		if !strings.HasSuffix(file.Name, fs.MDExt) {
			continue
		}
		if err := migrateFile(fsys, srcDir, file.Name, areaPath, kind, copy, dryRun, res); err != nil {
			return err
		}
	}

	for _, dir := range fs.OnlyDirs(entries) {
		if isLegacySkipDir(dir.Name) || isReservedAreaName(dir.Name) {
			continue
		}
		childSrc := fs.JoinDir(srcDir, dir.Name)
		sectionPath, created, err := ensureSection(fsys, areaPath, dir.Name, dryRun)
		if err != nil {
			return err
		}
		if created {
			res.SectionsCreated++
		}
		if err := migrateTree(fsys, childSrc, sectionPath, kind, copy, dryRun, res); err != nil {
			return err
		}
		if !copy && !dryRun {
			_ = fsys.DelDir(childSrc)
		}
	}
	return nil
}

func migrateFile(fsys *fs.FS, srcDir, filename, areaPath string, kind Kind, copy, dryRun bool, res *MigrateResult) error {
	docDir := DocDir(areaPath, kind)

	exists, err := fsys.Exists(docDir, filename)
	if err != nil {
		return err
	}
	if exists {
		res.FilesSkipped++
		return nil
	}

	if dryRun {
		res.FilesMoved++
		return nil
	}

	content, err := fsys.Read(srcDir, filename)
	if err != nil {
		return err
	}
	if err := fsys.Write(docDir, filename, content); err != nil {
		return err
	}
	if err := RegisterDoc(fsys, kind, docDir, filename); err != nil {
		return err
	}
	if !copy {
		if err := fsys.Del(srcDir, filename); err != nil {
			return err
		}
	}
	res.FilesMoved++
	return nil
}

// SphereForLegacyDir suggests a default sphere for a legacy folder name.
func SphereForLegacyDir(dirName string) string {
	key := strings.ToLower(strings.TrimSpace(dirName))
	switch {
	case strings.Contains(key, "work"), strings.Contains(key, "работ"), strings.Contains(key, "career"), strings.Contains(key, "карьер"), strings.Contains(key, "project"), strings.Contains(key, "проект"):
		return "Работа"
	case strings.Contains(key, "family"), strings.Contains(key, "friend"), strings.Contains(key, "семь"), strings.Contains(key, "друг"), strings.Contains(key, "отношен"):
		return "Отношения"
	case strings.Contains(key, "health"), strings.Contains(key, "sport"), strings.Contains(key, "здоров"), strings.Contains(key, "сон"), strings.Contains(key, "sleep"):
		return "Здоровье"
	case strings.Contains(key, "learn"), strings.Contains(key, "book"), strings.Contains(key, "курс"), strings.Contains(key, "обуч"), strings.Contains(key, "read"):
		return "Обучение"
	default:
		return "Личное"
	}
}
