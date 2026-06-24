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

// KindFromSubdir maps drafts/final/discussions folder names to Kind.
func KindFromSubdir(name string) (Kind, bool) {
	switch name {
	case SubDirDrafts:
		return KindDraft, true
	case SubDirFinal:
		return KindFinal, true
	case SubDirDiscussions:
		return KindDiscussion, true
	default:
		return KindDraft, false
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

// IsProjectPath reports spheres/<sphere>/<project> (top-level area only).
func IsProjectPath(path string) bool {
	parts := pathParts(path)
	return len(parts) == 3 && parts[0] == DirSpheres && !isReservedAreaName(parts[2])
}

// IsAreaPath reports any area directory under a sphere (any nesting depth).
func IsAreaPath(path string) bool {
	parts := pathParts(path)
	if len(parts) < 3 || parts[0] != DirSpheres {
		return false
	}
	return !isReservedAreaName(parts[len(parts)-1])
}

// IsDocDir reports an area's drafts/final/discussions folder.
func IsDocDir(dir string) bool {
	parts := pathParts(dir)
	if len(parts) < 4 || parts[0] != DirSpheres {
		return false
	}
	return isKindSubdir(parts[len(parts)-1])
}

// ProjectPathFromDoc returns the area path for a document directory.
func ProjectPathFromDoc(docDir string) (string, bool) {
	if !IsDocDir(docDir) {
		return "", false
	}
	parts := pathParts(docDir)
	return strings.Join(parts[:len(parts)-1], "/"), true
}

// AreaDepth returns nesting depth under a sphere (1 = top-level area).
func AreaDepth(areaPath string) int {
	parts := pathParts(areaPath)
	if len(parts) < 3 {
		return 0
	}
	return len(parts) - 2
}

// SpherePathFromArea returns the parent sphere path for an area.
func SpherePathFromArea(areaPath string) string {
	parts := pathParts(areaPath)
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts[:2], "/")
}

// ParentAreaPath returns the parent area or sphere for navigation back.
func ParentAreaPath(areaPath string) string {
	parts := pathParts(areaPath)
	if len(parts) <= 3 {
		return SpherePathFromArea(areaPath)
	}
	return strings.Join(parts[:len(parts)-1], "/")
}

// IsLifePath reports whether path is inside the life structure.
func IsLifePath(path string) bool {
	parts := pathParts(path)
	return len(parts) > 0 && parts[0] == DirSpheres
}

func isKindSubdir(name string) bool {
	return name == SubDirDrafts || name == SubDirFinal || name == SubDirDiscussions
}

func isReservedAreaName(name string) bool {
	return isKindSubdir(name) || name == SphereHubFile || name == ProjectHubFile || name == TasksFilename
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

// AreaTitle returns a display title for an area path (leaf name only).
func AreaTitle(projectPath string) string {
	return fs.DisplayName(baseName(projectPath))
}

// AreaFullTitle returns a slash-separated title for nested areas.
func AreaFullTitle(areaPath string) string {
	parts := pathParts(areaPath)
	if len(parts) < 3 {
		return ""
	}
	names := make([]string, 0, len(parts)-2)
	for _, name := range parts[2:] {
		names = append(names, fs.DisplayName(name))
	}
	return strings.Join(names, " / ")
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

// ListProjects returns top-level area paths inside a sphere.
func ListProjects(fsys *fs.FS, spherePath string) ([]string, error) {
	return ListChildAreas(fsys, spherePath)
}

// ListChildAreas returns immediate child area directories.
func ListChildAreas(fsys *fs.FS, parentPath string) ([]string, error) {
	entries, err := fsys.FilesAndDirs(parentPath)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, entry := range fs.OnlyDirs(entries) {
		if isReservedAreaName(entry.Name) {
			continue
		}
		paths = append(paths, fs.JoinDir(parentPath, entry.Name))
	}
	return paths, nil
}

// ListAllAreas returns every area under a sphere, depth-first.
func ListAllAreas(fsys *fs.FS, spherePath string) ([]string, error) {
	var areas []string
	if err := walkAreas(fsys, spherePath, func(path string) {
		areas = append(areas, path)
	}); err != nil {
		return nil, err
	}
	return areas, nil
}

func walkAreas(fsys *fs.FS, parentPath string, visit func(string)) error {
	children, err := ListChildAreas(fsys, parentPath)
	if err != nil {
		return err
	}
	for _, child := range children {
		visit(child)
		if err := walkAreas(fsys, child, visit); err != nil {
			return err
		}
	}
	return nil
}

// EnsureAreaDirs creates drafts/final/discussions under an area if missing.
func EnsureAreaDirs(fsys *fs.FS, areaPath string) error {
	for _, sub := range []string{SubDirDrafts, SubDirFinal, SubDirDiscussions} {
		if err := fsys.CreateDirsIfNotExist(fs.JoinDir(areaPath, sub)); err != nil {
			return err
		}
	}
	return nil
}

// CreateSection creates a nested area (раздел) inside a parent area or sphere.
func CreateSection(fsys *fs.FS, parentPath, sectionName string) (string, error) {
	sectionPath := fs.JoinDir(parentPath, fs.SanitizeFilename(sectionName))
	exists, err := fsys.Exists(sectionPath, "")
	if err != nil {
		return "", fmt.Errorf("create section: %w", err)
	}
	if exists {
		return sectionPath, nil
	}
	if err := fsys.MakeDir(sectionPath); err != nil {
		return "", fmt.Errorf("create section: %w", err)
	}
	if err := EnsureAreaDirs(fsys, sectionPath); err != nil {
		return "", err
	}
	displayTitle := fs.DisplayName(fs.Filename(sectionName))
	if err := fsys.Write(sectionPath, ProjectHubFile, projectHubTemplate(displayTitle)); err != nil {
		return "", fmt.Errorf("create section: %w", err)
	}
	return sectionPath, nil
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

// MoveDocKind moves a note between drafts/final/discussions within the same area.
func MoveDocKind(fsys *fs.FS, docDir, filename string, dstKind Kind) error {
	srcKind, ok := KindFromSubdir(baseName(docDir))
	if !ok {
		return fmt.Errorf("move doc kind: not a doc dir")
	}
	if srcKind == dstKind {
		return nil
	}
	projectPath, ok := ProjectPathFromDoc(docDir)
	if !ok {
		return fmt.Errorf("move doc kind: can't resolve area")
	}
	dstDir := DocDir(projectPath, dstKind)
	return relocateDoc(fsys, docDir, filename, dstDir)
}

// MoveDocToArea moves a note to the same status folder in another area.
func MoveDocToArea(fsys *fs.FS, docDir, filename, dstAreaPath string) error {
	srcKind, ok := KindFromSubdir(baseName(docDir))
	if !ok {
		return fmt.Errorf("move doc to area: not a doc dir")
	}
	srcArea, ok := ProjectPathFromDoc(docDir)
	if !ok {
		return fmt.Errorf("move doc to area: can't resolve area")
	}
	if srcArea == dstAreaPath {
		return nil
	}
	if err := EnsureAreaDirs(fsys, dstAreaPath); err != nil {
		return err
	}
	dstDir := DocDir(dstAreaPath, srcKind)
	return relocateDoc(fsys, docDir, filename, dstDir)
}

func relocateDoc(fsys *fs.FS, srcDocDir, filename, dstDocDir string) error {
	if srcDocDir == dstDocDir {
		return nil
	}
	srcKind, ok := KindFromSubdir(baseName(srcDocDir))
	if !ok {
		return fmt.Errorf("relocate doc: bad source dir")
	}
	dstKind, ok := KindFromSubdir(baseName(dstDocDir))
	if !ok {
		return fmt.Errorf("relocate doc: bad destination dir")
	}

	content, err := fsys.Read(srcDocDir, filename)
	if err != nil {
		return fmt.Errorf("relocate doc: can't read: %w", err)
	}
	if err := fsys.Write(dstDocDir, filename, content); err != nil {
		return fmt.Errorf("relocate doc: can't write: %w", err)
	}
	if err := fsys.Del(srcDocDir, filename); err != nil {
		return fmt.Errorf("relocate doc: can't delete source: %w", err)
	}

	oldLink := linkPath(srcDocDir, filename)
	newLink := linkPath(dstDocDir, filename)
	if err := unregisterDoc(fsys, srcKind, srcDocDir, filename); err != nil {
		return fmt.Errorf("relocate doc: can't unregister: %w", err)
	}
	if err := RegisterDoc(fsys, dstKind, dstDocDir, filename); err != nil {
		return fmt.Errorf("relocate doc: can't register: %w", err)
	}
	_ = oldLink
	_ = newLink
	return nil
}

func unregisterDoc(fsys *fs.FS, kind Kind, docDir, filename string) error {
	link := linkPath(docDir, filename)
	if err := removeLink(fsys, kind.section(), link); err != nil {
		return err
	}
	projectPath, ok := ProjectPathFromDoc(docDir)
	if !ok {
		return nil
	}
	return removeLinkFromFile(fsys, fs.JoinDir(projectPath, ProjectHubFile), kind.section(), link)
}

func removeLink(fsys *fs.FS, section, link string) error {
	return removeLinkFromFile(fsys, IndexFilename, section, link)
}

func removeLinkFromFile(fsys *fs.FS, hubFile, section, link string) error {
	dir := fs.DirUserRoot
	filename := hubFile
	if strings.Contains(hubFile, "/") {
		parts := pathParts(hubFile)
		filename = parts[len(parts)-1]
		dir = strings.Join(parts[:len(parts)-1], "/")
	}

	content, err := fsys.Read(dir, filename)
	if err != nil {
		return fmt.Errorf("remove link: can't read %s: %w", hubFile, err)
	}
	updated := strings.Replace(content, link+"\n", "", 1)
	if updated == content {
		updated = strings.Replace(content, link, "", 1)
	}
	return fsys.Write(dir, filename, updated)
}

// Finalize moves a draft to final/ inside the same project.
func Finalize(fsys *fs.FS, docDir, filename string) error {
	return MoveDocKind(fsys, docDir, filename, KindFinal)
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
	case IsAreaPath(srcPath):
		rel := strings.Join(parts[2:], "/")
		dstPath := fs.JoinDir(dstSpherePath, rel)
		if err := copyTree(fsys, srcPath, dstPath); err != nil {
			return err
		}
		return fsys.DelDir(srcPath)

	case IsDocDir(parentPath(srcPath)):
		docDir := parentPath(srcPath)
		filename := baseName(srcPath)
		areaPath, ok := ProjectPathFromDoc(docDir)
		if !ok {
			return fmt.Errorf("move: can't resolve area for doc")
		}
		kindDir := baseName(docDir)
		dstArea, err := ensureAreaInSphere(fsys, dstSpherePath, areaPath)
		if err != nil {
			return err
		}
		dstDir := fs.JoinDir(dstArea, kindDir)
		if err := fsys.CreateDirsIfNotExist(dstDir); err != nil {
			return err
		}
		content, err := fsys.Read(docDir, filename)
		if err != nil {
			return err
		}
		if err := fsys.Write(dstDir, filename, content); err != nil {
			return err
		}
		return fsys.Del(docDir, filename)

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

func ensureAreaInSphere(fsys *fs.FS, dstSpherePath, srcAreaPath string) (string, error) {
	parts := pathParts(srcAreaPath)
	if len(parts) < 3 {
		return "", fmt.Errorf("ensure area: bad path")
	}
	rel := strings.Join(parts[2:], "/")
	dstPath := fs.JoinDir(dstSpherePath, rel)
	exists, err := fsys.Exists(dstPath, "")
	if err != nil {
		return "", err
	}
	if exists {
		return dstPath, nil
	}
	parent := dstSpherePath
	for _, name := range parts[2:] {
		child, err := CreateSection(fsys, parent, name)
		if err != nil {
			return "", err
		}
		parent = child
	}
	return dstPath, nil
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
