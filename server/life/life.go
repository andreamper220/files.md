// Package life defines spheres, projects, and document kinds on disk.
// A sphere is spheres/<Name>/; a project is a first-level subfolder inside a sphere.
package life

import (
	"fmt"
	"strings"

	"github.com/zakirullin/files.md/server/fs"
)

const (
	DirSpheres        = "spheres"
	SubDirDrafts      = "drafts"
	SubDirFinal       = "final"
	SubDirDiscussions = "discussions"
	SphereHubFile     = "Sphere.md"
	ProjectHubFile    = "Project.md"
	IndexFilename     = "Life.md"
)

const (
	sectionSpheres     = "## Сферы"
	sectionProjects    = "## Проекты"
	sectionDrafts      = "## Черновики"
	sectionFinal       = "## Финальные"
	sectionDiscussions = "## Обсуждения"
)

var defaultSpheres = []struct {
	dirName string
	title   string
	about   string
}{
	{dirName: "Работа", title: "Работа", about: "карьера, проекты, профессиональное развитие"},
	{dirName: "Отношения", title: "Отношения", about: "семья, друзья, близкие люди"},
	{dirName: "Здоровье", title: "Здоровье", about: "тело, сон, спорт, питание"},
	{dirName: "Личное", title: "Личное", about: "внутренний мир, хобби, отдых"},
	{dirName: "Обучение", title: "Обучение", about: "книги, курсы, навыки"},
}

// legacySphereTitles maps old English folder names to Russian display titles.
var legacySphereTitles = map[string]string{
	"Work":          "Работа",
	"Relationships": "Отношения",
	"Health":        "Здоровье",
	"Personal":      "Личное",
	"Learning":      "Обучение",
}

// Kind marks a document category inside a project.
type Kind int

const (
	KindDraft Kind = iota
	KindFinal
	KindDiscussion
)

func (k Kind) section() string {
	switch k {
	case KindDraft:
		return sectionDrafts
	case KindFinal:
		return sectionFinal
	case KindDiscussion:
		return sectionDiscussions
	default:
		return sectionDrafts
	}
}

func (k Kind) subdir() string {
	switch k {
	case KindDraft:
		return SubDirDrafts
	case KindFinal:
		return SubDirFinal
	case KindDiscussion:
		return SubDirDiscussions
	default:
		return SubDirDrafts
	}
}

// KindFromCode parses a one-letter kind code from callback data.
func KindFromCode(code string) (Kind, bool) {
	switch code {
	case "d":
		return KindDraft, true
	case "f":
		return KindFinal, true
	case "c":
		return KindDiscussion, true
	default:
		return KindDraft, false
	}
}

// KindCode returns a one-letter code for callback data.
func KindCode(k Kind) string {
	switch k {
	case KindDraft:
		return "d"
	case KindFinal:
		return "f"
	case KindDiscussion:
		return "c"
	default:
		return "d"
	}
}

// DocDir returns the directory for a document kind inside a project.
func DocDir(projectPath string, kind Kind) string {
	return fs.JoinDir(projectPath, kind.subdir())
}

// SpherePath builds spheres/<name>.
func SpherePath(name string) string {
	return fs.JoinDir(DirSpheres, fs.SanitizeFilename(name))
}

// ProjectPath builds spheres/<sphere>/<project>.
func ProjectPath(spherePath, projectName string) string {
	return fs.JoinDir(spherePath, fs.SanitizeFilename(projectName))
}

// IsSpherePath reports spheres/<sphere>.
func IsSpherePath(path string) bool {
	parts := pathParts(path)
	return len(parts) == 2 && parts[0] == DirSpheres
}

// IsProjectPath reports spheres/<sphere>/<project>.
func IsProjectPath(path string) bool {
	parts := pathParts(path)
	return len(parts) == 3 && parts[0] == DirSpheres && !isKindSubdir(parts[2])
}

// IsDocDir reports a project's drafts/final/discussions folder.
func IsDocDir(dir string) bool {
	parts := pathParts(dir)
	if len(parts) != 4 || parts[0] != DirSpheres {
		return false
	}
	return isKindSubdir(parts[3])
}

// ProjectPathFromDoc returns the project path for a document directory.
func ProjectPathFromDoc(docDir string) (string, bool) {
	if !IsDocDir(docDir) {
		return "", false
	}
	parts := pathParts(docDir)
	return strings.Join(parts[:3], "/"), true
}

// IsLifePath reports whether path is inside the life structure.
func IsLifePath(path string) bool {
	parts := pathParts(path)
	return len(parts) > 0 && parts[0] == DirSpheres
}

func isKindSubdir(name string) bool {
	return name == SubDirDrafts || name == SubDirFinal || name == SubDirDiscussions
}

func pathParts(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

func baseName(path string) string {
	parts := pathParts(path)
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func parentPath(path string) string {
	parts := pathParts(path)
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "/")
}

// EnsureSpheresRoot creates the spheres/ directory.
func EnsureSpheresRoot(fsys *fs.FS) error {
	return fsys.CreateDirsIfNotExist(DirSpheres)
}

// Init creates Life.md and default sphere folders.
func Init(fsys *fs.FS) error {
	if err := EnsureSpheresRoot(fsys); err != nil {
		return fmt.Errorf("life init: can't create spheres root: %w", err)
	}

	exists, err := fsys.Exists(fs.DirUserRoot, IndexFilename)
	if err != nil {
		return fmt.Errorf("life init: can't check index: %w", err)
	}
	if !exists {
		if err := fsys.Write(fs.DirUserRoot, IndexFilename, indexTemplate()); err != nil {
			return fmt.Errorf("life init: can't write index: %w", err)
		}
	}

	for _, sphere := range defaultSpheres {
		spherePath := SpherePath(sphere.dirName)
		exists, err := fsys.Exists(spherePath, "")
		if err != nil {
			return fmt.Errorf("life init: can't check sphere %s: %w", sphere.dirName, err)
		}
		if exists {
			continue
		}
		if err := createSphere(fsys, spherePath, sphere.title, sphere.about); err != nil {
			return fmt.Errorf("life init: can't create sphere %s: %w", sphere.dirName, err)
		}
	}

	return nil
}

func createSphere(fsys *fs.FS, spherePath, title, about string) error {
	if err := fsys.MakeDir(spherePath); err != nil {
		return err
	}
	body := sphereHubTemplate(title, about)
	if err := fsys.Write(spherePath, SphereHubFile, body); err != nil {
		return err
	}
	link := linkWithLabel(title, spherePath, SphereHubFile)
	return appendLink(fsys, sectionSpheres, link)
}

// SphereTitle returns a Russian display title for a sphere path.
func SphereTitle(spherePath string) string {
	name := baseName(spherePath)
	for _, sphere := range defaultSpheres {
		if sphere.dirName == name || sphere.title == name {
			return sphere.title
		}
	}
	if title, ok := legacySphereTitles[name]; ok {
		return title
	}
	return fs.DisplayName(name)
}

// IsInitialized reports whether the life structure has been set up.
func IsInitialized(fsys *fs.FS) bool {
	exists, err := fsys.Exists(fs.DirUserRoot, IndexFilename)
	if err != nil || !exists {
		return false
	}
	spheres, err := ListSpheres(fsys)
	return err == nil && len(spheres) > 0
}

// CreateProject creates a project folder inside a sphere with document subdirs.
func CreateProject(fsys *fs.FS, spherePath, projectName string) (string, error) {
	projectPath := ProjectPath(spherePath, projectName)
	exists, err := fsys.Exists(projectPath, "")
	if err != nil {
		return "", fmt.Errorf("create project: %w", err)
	}
	if exists {
		return projectPath, nil
	}

	if err := fsys.MakeDir(projectPath); err != nil {
		return "", fmt.Errorf("create project: can't make dir: %w", err)
	}
	for _, sub := range []string{SubDirDrafts, SubDirFinal, SubDirDiscussions} {
		if err := fsys.MakeDir(fs.JoinDir(projectPath, sub)); err != nil {
			return "", fmt.Errorf("create project: can't make %s: %w", sub, err)
		}
	}

	displayTitle := fs.DisplayName(fs.Filename(projectName))
	if err := fsys.Write(projectPath, ProjectHubFile, projectHubTemplate(displayTitle)); err != nil {
		return "", fmt.Errorf("create project: can't write hub: %w", err)
	}

	projectLink := linkWithLabel(fs.DisplayName(baseName(projectPath)), projectPath, ProjectHubFile)
	if err := appendLink(fsys, sectionProjects, projectLink); err != nil {
		return "", err
	}
	if err := appendLinkToFile(fsys, fs.JoinDir(spherePath, SphereHubFile), "## Проекты", projectLink); err != nil {
		return "", err
	}

	return projectPath, nil
}

// ListSpheres returns sphere directory paths under spheres/.
func ListSpheres(fsys *fs.FS) ([]string, error) {
	entries, err := fsys.FilesAndDirs(DirSpheres)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, entry := range fs.OnlyDirs(entries) {
		paths = append(paths, fs.JoinDir(DirSpheres, entry.Name))
	}
	return paths, nil
}

// ListProjects returns project directory paths inside a sphere.
func ListProjects(fsys *fs.FS, spherePath string) ([]string, error) {
	entries, err := fsys.FilesAndDirs(spherePath)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, entry := range fs.OnlyDirs(entries) {
		if isKindSubdir(entry.Name) {
			continue
		}
		paths = append(paths, fs.JoinDir(spherePath, entry.Name))
	}
	return paths, nil
}

// RegisterDoc adds a link to Life.md and the project hub.
func RegisterDoc(fsys *fs.FS, kind Kind, docDir, filename string) error {
	link := linkPath(docDir, filename)
	if err := appendLink(fsys, kind.section(), link); err != nil {
		return err
	}
	projectPath, ok := ProjectPathFromDoc(docDir)
	if !ok {
		return nil
	}
	return appendLinkToFile(fsys, fs.JoinDir(projectPath, ProjectHubFile), kind.section(), link)
}

// Finalize moves a draft to final/ inside the same project.
func Finalize(fsys *fs.FS, docDir, filename string) error {
	if baseName(docDir) != SubDirDrafts {
		return fmt.Errorf("finalize: not a draft")
	}

	projectPath, ok := ProjectPathFromDoc(docDir)
	if !ok {
		return fmt.Errorf("finalize: can't resolve project")
	}

	content, err := fsys.Read(docDir, filename)
	if err != nil {
		return fmt.Errorf("finalize: can't read draft: %w", err)
	}
	finalDir := DocDir(projectPath, KindFinal)
	if err := fsys.Write(finalDir, filename, content); err != nil {
		return fmt.Errorf("finalize: can't write: %w", err)
	}
	if err := fsys.Del(docDir, filename); err != nil {
		return fmt.Errorf("finalize: can't delete draft: %w", err)
	}

	oldLink := linkPath(docDir, filename)
	newLink := linkPath(finalDir, filename)
	if err := moveLink(fsys, sectionDrafts, sectionFinal, oldLink, newLink); err != nil {
		return fmt.Errorf("finalize: can't update life index: %w", err)
	}
	hubFile := fs.JoinDir(projectPath, ProjectHubFile)
	if err := moveLinkInFile(fsys, hubFile, sectionDrafts, sectionFinal, oldLink, newLink); err != nil {
		return fmt.Errorf("finalize: can't update project hub: %w", err)
	}

	return nil
}

// MoveEntry moves a project or file to another sphere.
func MoveEntry(fsys *fs.FS, srcPath, dstSpherePath string) error {
	if !IsLifePath(srcPath) {
		return fmt.Errorf("move: not a life path")
	}
	if !IsSpherePath(dstSpherePath) {
		return fmt.Errorf("move: target is not a sphere")
	}

	srcSphere := strings.Join(pathParts(srcPath)[:2], "/")
	if strings.HasPrefix(srcPath, dstSpherePath+"/") || srcSphere == dstSpherePath {
		return nil
	}

	parts := pathParts(srcPath)

	switch {
	case IsProjectPath(srcPath):
		projectName := parts[2]
		dstPath := ProjectPath(dstSpherePath, projectName)
		if err := copyTree(fsys, srcPath, dstPath); err != nil {
			return err
		}
		return fsys.DelDir(srcPath)

	case len(parts) == 5:
		projectName := parts[2]
		kindDir := parts[3]
		filename := parts[4]
		dstProject, err := ensureProjectInSphere(fsys, dstSpherePath, projectName)
		if err != nil {
			return err
		}
		dstDir := fs.JoinDir(dstProject, kindDir)
		if err := fsys.CreateDirsIfNotExist(dstDir); err != nil {
			return err
		}
		content, err := fsys.Read(strings.Join(parts[:4], "/"), filename)
		if err != nil {
			return err
		}
		if err := fsys.Write(dstDir, filename, content); err != nil {
			return err
		}
		return fsys.Del(strings.Join(parts[:4], "/"), filename)

	case len(parts) == 3 && strings.HasSuffix(srcPath, fs.MDExt):
		filename := parts[2]
		srcSphereDir := strings.Join(parts[:2], "/")
		content, err := fsys.Read(srcSphereDir, filename)
		if err != nil {
			return err
		}
		if err := fsys.Write(dstSpherePath, filename, content); err != nil {
			return err
		}
		return fsys.Del(srcSphereDir, filename)

	default:
		return fmt.Errorf("move: unsupported path %s", srcPath)
	}
}

func copyTree(fsys *fs.FS, srcPath, dstPath string) error {
	if err := fsys.MakeDir(dstPath); err != nil {
		return err
	}
	entries, err := fsys.FilesAndDirs(srcPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir {
			if err := copyTree(fsys, fs.JoinDir(srcPath, entry.Name), fs.JoinDir(dstPath, entry.Name)); err != nil {
				return err
			}
			continue
		}
		content, err := fsys.Read(srcPath, entry.Name)
		if err != nil {
			return err
		}
		if err := fsys.Write(dstPath, entry.Name, content); err != nil {
			return err
		}
	}
	return nil
}

func ensureProjectInSphere(fsys *fs.FS, spherePath, projectName string) (string, error) {
	projectPath := ProjectPath(spherePath, projectName)
	exists, err := fsys.Exists(projectPath, "")
	if err != nil {
		return "", err
	}
	if exists {
		return projectPath, nil
	}
	return CreateProject(fsys, spherePath, projectName)
}

// ShortIndexText returns the life index without the header and sphere list (for the bot).
func ShortIndexText(fsys *fs.FS) (string, error) {
	if !IsInitialized(fsys) {
		return "Структура жизни ещё не создана.", nil
	}

	content, err := fsys.Read(fs.DirUserRoot, IndexFilename)
	if err != nil {
		return "", fmt.Errorf("life index: can't read: %w", err)
	}

	if idx := strings.Index(content, sectionProjects); idx != -1 {
		content = strings.TrimSpace(content[idx:])
	} else {
		content = sectionProjects + "\n\n" + sectionDrafts + "\n\n" + sectionFinal + "\n\n" + sectionDiscussions
	}

	const maxRunes = 3500
	runes := []rune(content)
	if len(runes) > maxRunes {
		content = string(runes[:maxRunes]) + "\n\n…"
	}
	return content, nil
}

// IndexPreview returns Life.md content for the bot.
func IndexPreview(fsys *fs.FS) (string, error) {
	return ShortIndexText(fsys)
}

func indexTemplate() string {
	return `## Проекты

## Черновики

## Финальные

## Обсуждения
`
}

func sphereHubTemplate(title, about string) string {
	return fmt.Sprintf(`# %s

> Сфера: %s

## Сейчас

## Проекты

## Заметки
`, title, about)
}

func projectHubTemplate(title string) string {
	return fmt.Sprintf(`# %s

> Проект

## Статус

## Черновики

## Финальные

## Обсуждения

## Задачи

## Заметки
`, title)
}

func linkPath(dir, filename string) string {
	if filename == "" {
		return fmt.Sprintf("- [%s](%s)", fs.DisplayName(baseName(dir)), dir)
	}
	label := fs.DisplayName(filename)
	if filename == SphereHubFile {
		label = SphereTitle(dir)
	}
	if filename == ProjectHubFile {
		label = fs.DisplayName(baseName(dir))
	}
	return fmt.Sprintf("- [%s](%s/%s)", label, dir, filename)
}

func linkWithLabel(label, dir, filename string) string {
	return fmt.Sprintf("- [%s](%s/%s)", label, dir, filename)
}

func appendLink(fsys *fs.FS, section, link string) error {
	return appendLinkToFile(fsys, IndexFilename, section, link)
}

func appendLinkToFile(fsys *fs.FS, hubFile, section, link string) error {
	dir := fs.DirUserRoot
	filename := hubFile
	if strings.Contains(hubFile, "/") {
		parts := pathParts(hubFile)
		filename = parts[len(parts)-1]
		dir = strings.Join(parts[:len(parts)-1], "/")
	}

	content, err := fsys.Read(dir, filename)
	if err != nil {
		return fmt.Errorf("append link: can't read %s: %w", hubFile, err)
	}
	if strings.Contains(content, link) {
		return nil
	}

	updated, ok := insertAfterSection(content, section, link+"\n")
	if !ok {
		updated = strings.TrimRight(content, "\n") + "\n\n" + section + "\n" + link + "\n"
	}

	return fsys.Write(dir, filename, updated)
}

func moveLink(fsys *fs.FS, fromSection, toSection, oldLink, newLink string) error {
	return moveLinkInFile(fsys, IndexFilename, fromSection, toSection, oldLink, newLink)
}

func moveLinkInFile(fsys *fs.FS, hubFile, fromSection, toSection, oldLink, newLink string) error {
	dir := fs.DirUserRoot
	filename := hubFile
	if strings.Contains(hubFile, "/") {
		parts := pathParts(hubFile)
		filename = parts[len(parts)-1]
		dir = strings.Join(parts[:len(parts)-1], "/")
	}

	content, err := fsys.Read(dir, filename)
	if err != nil {
		return fmt.Errorf("move link: can't read: %w", err)
	}

	content = strings.Replace(content, oldLink, "", 1)
	content = strings.TrimRight(content, "\n") + "\n"

	updated, ok := insertAfterSection(content, toSection, newLink+"\n")
	if !ok {
		updated = content + "\n" + toSection + "\n" + newLink + "\n"
	}
	_ = fromSection
	return fsys.Write(dir, filename, updated)
}

func insertAfterSection(content, section, insert string) (string, bool) {
	idx := strings.Index(content, section)
	if idx == -1 {
		return "", false
	}

	after := content[idx+len(section):]
	nextHeader := strings.Index(after, "\n## ")
	var tail string
	if nextHeader == -1 {
		tail = after
		after = ""
	} else {
		tail = after[:nextHeader]
		after = after[nextHeader:]
	}

	if strings.Contains(tail, strings.TrimSpace(insert)) {
		return content, true
	}

	newTail := strings.TrimRight(tail, "\n") + "\n" + insert
	return content[:idx] + section + newTail + after, true
}
