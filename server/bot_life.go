package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/i18n"
	"github.com/zakirullin/files.md/server/life"
	"github.com/zakirullin/files.md/server/pkg/tg"
)

const CmdLifeRecentProject = "life_prj_recent"

func (b *Bot) initLife(_ []string) error {
	if err := life.Init(b.fs); err != nil {
		return fmt.Errorf("init life: %w", err)
	}

	kb := tg.NewKeyboard([]tg.Row{
		tg.NewBtn(i18n.Tr("🌐 Сферы"), tg.NewCmd(CmdShowLifeSpheres, nil)),
		tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)),
	})
	return b.showHTML(i18n.Tr("Структура жизни создана 👌\n\nОткрой <b>Life.md</b> в приложении."), kb)
}

func (b *Bot) showLifeSpheres(_ []string) error {
	_ = life.EnsureSpheresRoot(b.fs)
	spheres, err := life.ListSpheres(b.fs)
	if err != nil {
		return fmt.Errorf("show life spheres: %w", err)
	}

	var kb tg.Keyboard
	for _, spherePath := range spheres {
		btn := tg.NewBtn(
			life.SphereTitle(spherePath),
			tg.NewCmd(CmdShowLifeSphere, []string{fs.ShortHash(spherePath)}),
		)
		kb.AddRow(btn)
	}
	if len(spheres) == 0 {
		kb.AddRow(tg.NewBtn(i18n.Tr("🏗 Создать структуру"), tg.NewCmd(CmdInitLife, nil)))
	}
	kb.AddRow(tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)))

	return b.showHTML(i18n.Tr("🌐 Сферы жизни:"), &kb)
}

func (b *Bot) showLifeSphere(params []string) error {
	spherePath, err := b.fs.ResolveDirParam(params[0])
	if err != nil {
		return fmt.Errorf("show life sphere: %w", err)
	}

	projects, err := life.ListProjects(b.fs, spherePath)
	if err != nil {
		return fmt.Errorf("show life sphere: can't list projects: %w", err)
	}

	var kb tg.Keyboard
	for _, projectPath := range projects {
		btn := tg.NewBtn(
			life.NestedAreaLabel(projectPath),
			tg.NewCmd(CmdShowLifeProject, []string{fs.ShortHash(projectPath)}),
		)
		kb.AddRow(btn)
	}

	kb.AddRow(tg.NewBtn(i18n.Tr("➕ Новая область"), tg.NewCmd(CmdLifeNewProject, []string{fs.ShortHash(spherePath)})))
	kb.AddRow(tg.NewRow(
		tg.NewBtn(i18n.Tr("🌐 Сферы"), tg.NewCmd(CmdShowLifeSpheres, nil)),
		tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)),
	))

	title := fmt.Sprintf("%s %s", i18n.Tr("🏗 Области сферы:"), life.SphereTitle(spherePath))
	return b.showHTML(title, &kb)
}

func (b *Bot) showLifeProject(params []string) error {
	projectPath, err := b.fs.ResolveDirParam(params[0])
	if err != nil {
		return fmt.Errorf("show life project: %w", err)
	}

	var kb tg.Keyboard
	sections, _ := life.ListChildAreas(b.fs, projectPath)
	for _, sectionPath := range sections {
		btn := tg.NewBtn(
			life.NestedAreaLabel(sectionPath),
			tg.NewCmd(CmdShowLifeProject, []string{fs.ShortHash(sectionPath)}),
		)
		kb.AddRow(btn)
	}

	kb.AddRow(tg.NewRow(
		tg.NewBtn(i18n.Tr("📝 Черновики"), tg.NewCmd(CmdShowLifeDocs, []string{fs.ShortHash(projectPath), life.KindCode(life.KindDraft)})),
		tg.NewBtn(i18n.Tr("✨ Финальные"), tg.NewCmd(CmdShowLifeDocs, []string{fs.ShortHash(projectPath), life.KindCode(life.KindFinal)})),
	))
	kb.AddRow(tg.NewBtn(i18n.Tr("💬 Обсуждения"), tg.NewCmd(CmdShowLifeDocs, []string{fs.ShortHash(projectPath), life.KindCode(life.KindDiscussion)})))
	kb.AddRow(tg.NewBtn(i18n.Tr("📄 Проект"), tg.NewCmd(CmdShowFile, []string{fs.ShortHash(projectPath), fs.ShortHash(life.ProjectHubFile)})))

	if life.IsAreaPath(projectPath) {
		kb.AddRow(tg.NewBtn("➕", tg.NewCmd(CmdLifeNewSection, []string{fs.ShortHash(projectPath)})))
		kb.AddRow(tg.NewBtn("↔️", tg.NewCmd(CmdShowMoveToSphere, []string{fs.ShortHash(projectPath)})))
	}

	parent := life.ParentAreaPath(projectPath)
	if life.IsSpherePath(parent) {
		kb.AddRow(tg.NewBtn("⬅️", tg.NewCmd(CmdShowLifeSphere, []string{fs.ShortHash(parent)})))
	} else if parent != "" {
		kb.AddRow(tg.NewBtn("⬅️", tg.NewCmd(CmdShowLifeProject, []string{fs.ShortHash(parent)})))
	}
	kb.AddRow(tg.NewBtn("🌐", tg.NewCmd(CmdShowLifeSpheres, nil)))

	kb.AddRow(tg.NewBtn("🏠", tg.NewCmd(CmdShowHome, nil)))

	title := fmt.Sprintf("%s %s", i18n.Tr("🏗 Область:"), life.AreaFullTitle(projectPath))
	return b.showHTML(title, &kb)
}

func (b *Bot) showLifeDocs(params []string) error {
	projectPath, err := b.fs.ResolveDirParam(params[0])
	if err != nil {
		return fmt.Errorf("show life docs: %w", err)
	}
	kind, ok := life.KindFromCode(params[1])
	if !ok {
		return fmt.Errorf("show life docs: unknown kind")
	}

	docDir := life.DocDir(projectPath, kind)
	entries, err := b.fs.FilesAndDirs(docDir)
	if err != nil {
		return fmt.Errorf("show life docs: %w", err)
	}

	var kb tg.Keyboard
	for _, file := range fs.SortByCtimeDesc(fs.OnlyFiles(entries)) {
		if !strings.HasSuffix(file.Name, fs.MDExt) {
			continue
		}
		kb.AddRow(tg.NewBtn(
			fs.DisplayName(file.Name),
			tg.NewCmd(CmdShowFile, []string{fs.ShortHash(docDir), fs.ShortHash(file.Name)}),
		))
	}
	if len(entries) == 0 {
		kb.AddRow(tg.NewBtn(i18n.Tr("Пока пусто"), tg.NewCmd(CmdDoNothing, nil)))
	}

	kb.AddRow(tg.NewRow(
		tg.NewBtn(i18n.Tr("🏗 Проект"), tg.NewCmd(CmdShowLifeProject, []string{fs.ShortHash(projectPath)})),
		tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)),
	))

	return b.showHTML(lifeKindLabel(kind), &kb)
}

func (b *Bot) moveToDraft(params []string) error {
	return b.showLifeSpherePicker(life.KindDraft, params[0])
}

func (b *Bot) moveToFinalize(params []string) error {
	return b.showLifeSpherePicker(life.KindFinal, params[0])
}

func (b *Bot) moveToDiscussion(params []string) error {
	return b.showLifeSpherePicker(life.KindDiscussion, params[0])
}

func (b *Bot) showLifeSpherePicker(kind life.Kind, msgHash string) error {
	_ = life.EnsureSpheresRoot(b.fs)
	spheres, err := life.ListSpheres(b.fs)
	if err != nil {
		return fmt.Errorf("life sphere picker: %w", err)
	}

	var kb tg.Keyboard
	kindCode := life.KindCode(kind)
	for _, spherePath := range spheres {
		kb.AddRow(tg.NewBtn(
			life.SphereTitle(spherePath),
			tg.NewCmd(CmdLifePickProject, []string{fs.ShortHash(spherePath), kindCode, msgHash}),
		))
	}
	if len(spheres) == 0 {
		kb.AddRow(tg.NewBtn(i18n.Tr("🏗 Создать структуру"), tg.NewCmd(CmdInitLife, nil)))
	}

	return b.showHTML(i18n.Tr("Выбери сферу:"), &kb)
}

func (b *Bot) showLifeProjectPicker(params []string) error {
	spherePath, err := b.fs.ResolveDirParam(params[0])
	if err != nil {
		return fmt.Errorf("life project picker: %w", err)
	}
	kind, ok := life.KindFromCode(params[1])
	if !ok {
		return fmt.Errorf("life project picker: bad kind")
	}
	msgHash := params[2]

	projects, err := life.ListProjects(b.fs, spherePath)
	if err != nil {
		return fmt.Errorf("life project picker: %w", err)
	}

	var kb tg.Keyboard
	for _, projectPath := range projects {
		kb.AddRow(tg.NewBtn(
			fs.DisplayName(baseName(projectPath)),
			tg.NewCmd(CmdLifeSaveToProject, []string{fs.ShortHash(projectPath), life.KindCode(kind), msgHash}),
		))
	}

	b.db.SetInputExpectation(tg.NewCmd(CmdLifeCreateProject, []string{fs.ShortHash(spherePath), life.KindCode(kind), msgHash, "%s"}))
	kb.AddRow(tg.NewBtn(i18n.Tr("➕ Новый проект"), tg.NewCmd(CmdDoNothing, nil)))

	return b.showHTML(i18n.Tr("Выбери проект или пришли имя нового:"), &kb)
}

func (b *Bot) saveToLifeProject(params []string) error {
	projectPath, err := b.fs.ResolveDirParam(params[0])
	if err != nil {
		return fmt.Errorf("save to project: %w", err)
	}
	kind, ok := life.KindFromCode(params[1])
	if !ok {
		return fmt.Errorf("save to project: bad kind")
	}
	msgHash := params[2]

	return b.saveChatToLifeProject(projectPath, kind, msgHash)
}

func (b *Bot) createLifeProject(params []string) error {
	spherePath, err := b.fs.ResolveDirParam(params[0])
	if err != nil {
		return fmt.Errorf("create life project: %w", err)
	}
	kind, ok := life.KindFromCode(params[1])
	if !ok {
		return fmt.Errorf("create life project: bad kind")
	}
	msgHash := params[2]
	projectName := params[3]

	projectPath, err := life.CreateProject(b.fs, spherePath, projectName)
	if err != nil {
		return fmt.Errorf("create life project: %w", err)
	}

	return b.saveChatToLifeProject(projectPath, kind, msgHash)
}

func (b *Bot) lifeNewProject(params []string) error {
	sphereHash := params[0]
	b.db.SetInputExpectation(tg.NewCmd(CmdLifeCreateProjectOnly, []string{sphereHash, "%s"}))
	return b.showHTML(i18n.Tr("Пришли имя нового проекта:"), nil)
}

func (b *Bot) createLifeProjectOnly(params []string) error {
	spherePath, err := b.fs.ResolveDirParam(params[0])
	if err != nil {
		return fmt.Errorf("create project only: %w", err)
	}
	projectName := params[1]

	projectPath, err := life.CreateProject(b.fs, spherePath, projectName)
	if err != nil {
		return fmt.Errorf("create project only: %w", err)
	}

	kb := tg.NewKeyboard([]tg.Row{
		tg.NewBtn(i18n.Tr("🏗 Проект"), tg.NewCmd(CmdShowLifeProject, []string{fs.ShortHash(projectPath)})),
		tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)),
	})
	msg := fmt.Sprintf(i18n.Tr("Проект <b>%s</b> создан"), life.AreaFullTitle(projectPath))
	return b.showHTML(msg, kb)
}

func (b *Bot) lifeNewSection(params []string) error {
	areaHash := params[0]
	b.db.SetInputExpectation(tg.NewCmd(CmdLifeCreateSection, []string{areaHash, "%s"}))
	return b.showHTML(i18n.Tr("Пришли имя нового раздела:"), nil)
}

func (b *Bot) createLifeSection(params []string) error {
	parentPath, err := b.fs.ResolveDirParam(params[0])
	if err != nil {
		return fmt.Errorf("create section: %w", err)
	}
	sectionName := params[1]

	sectionPath, err := life.CreateSection(b.fs, parentPath, sectionName)
	if err != nil {
		return fmt.Errorf("create section: %w", err)
	}

	kb := tg.NewKeyboard([]tg.Row{
		tg.NewBtn("🏗", tg.NewCmd(CmdShowLifeProject, []string{fs.ShortHash(sectionPath)})),
		tg.NewBtn("🏠", tg.NewCmd(CmdShowHome, nil)),
	})
	msg := fmt.Sprintf(i18n.Tr("Раздел <b>%s</b> создан"), life.AreaFullTitle(sectionPath))
	return b.showHTML(msg, kb)
}

func (b *Bot) saveChatToLifeProject(projectPath string, kind life.Kind, msgHash string) error {
	docDir := life.DocDir(projectPath, kind)

	var savedFilename string
	err := b.moveFromChat(func(content string, timestamp time.Time) error {
		var extractErr error
		var sanitizedTitle string
		sanitizedTitle, content, extractErr = b.extractHeaderAndBodyPreserveMedia(content, maxHeaderLength)
		if extractErr != nil {
			return fmt.Errorf("save to project: can't extract title: %w", extractErr)
		}

		savedFilename = fs.Filename(sanitizedTitle)
		if err := b.createOrAdd(docDir, savedFilename, content); err != nil {
			return err
		}
		return life.RegisterDoc(b.fs, kind, docDir, savedFilename)
	}, true, msgHash)
	if err != nil {
		return fmt.Errorf("save to project: %w", err)
	}

	b.setRecentLifeProject(projectPath)
	b.delAllKeyboards()

	label := lifeKindLabel(kind)
	spherePath := life.SpherePathFromArea(projectPath)
	projectLabel := life.SaveLocationLabel(spherePath, projectPath)
	msg := fmt.Sprintf(i18n.Tr("Сохранено в <b>%s</b> → %s"), projectLabel, label)
	_, _ = b.tg.Send(b.userID, msg, nil, tg.MarkupHTML)

	return b.ShowHome(nil)
}

func (b *Bot) addToLifeFromShortcut(kind life.Kind, params []string) error {
	content := params[0]
	projectPath, rest, ok := parseLifeShortcutTarget(content)
	if ok {
		content = rest
	} else {
		projectPath = b.recentLifeProject()
		if projectPath == "" {
			_, _ = b.tg.Send(b.userID, i18n.Tr("Укажи проект: <code>д Сфера/Проект текст</code> или сначала сохрани через кнопки."), nil, tg.MarkupHTML)
			return b.ShowHome(nil)
		}
	}

	spherePath, fullProjectPath, err := resolveLifeShortcutProject(b.fs, projectPath)
	if err != nil {
		return fmt.Errorf("life shortcut: %w", err)
	}
	_ = spherePath
	if !life.IsAreaPath(fullProjectPath) {
		return fmt.Errorf("life shortcut: can't resolve area")
	}

	title, body, err := b.extractHeaderAndBody(content, maxHeaderLength)
	if err != nil {
		return fmt.Errorf("life shortcut: %w", err)
	}

	docDir := life.DocDir(fullProjectPath, kind)
	filename := fs.Filename(title)
	if err := b.createOrAdd(docDir, filename, body); err != nil {
		return fmt.Errorf("life shortcut: can't save: %w", err)
	}
	if err := life.RegisterDoc(b.fs, kind, docDir, filename); err != nil {
		return fmt.Errorf("life shortcut: can't register: %w", err)
	}

	b.setRecentLifeProject(fullProjectPath)
	label := lifeKindLabel(kind)
	msg := fmt.Sprintf(i18n.Tr("Сохранено в <b>%s</b> → %s"), life.SaveLocationLabel(life.SpherePathFromArea(fullProjectPath), fullProjectPath), label)
	_, _ = b.tg.Send(b.userID, msg, nil, tg.MarkupHTML)

	return b.ShowHome(nil)
}

func (b *Bot) addToDraftFromShortcut(params []string) error {
	return b.addToLifeFromShortcut(life.KindDraft, params)
}

func (b *Bot) addToFinalizeFromShortcut(params []string) error {
	return b.addToLifeFromShortcut(life.KindFinal, params)
}

func (b *Bot) addToDiscussionFromShortcut(params []string) error {
	return b.addToLifeFromShortcut(life.KindDiscussion, params)
}

func (b *Bot) finalizeDoc(params []string) error {
	dirHash := params[0]
	filenameHash := params[1]

	dir, err := b.fs.ResolveDirParam(dirHash)
	if err != nil {
		return fmt.Errorf("finalize: can't resolve dir: %w", err)
	}

	filename, err := b.fs.Unhash(dir, filenameHash)
	if err != nil {
		return fmt.Errorf("finalize: can't resolve file: %w", err)
	}

	if err := life.Finalize(b.fs, dir, filename); err != nil {
		return fmt.Errorf("finalize: %w", err)
	}

	projectPath, _ := life.ProjectPathFromDoc(dir)
	kb := tg.NewKeyboard([]tg.Row{
		tg.NewBtn(i18n.Tr("✨ Финальные"), tg.NewCmd(CmdShowLifeDocs, []string{fs.ShortHash(projectPath), life.KindCode(life.KindFinal)})),
		tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)),
	})
	return b.showHTML(i18n.Tr("Финализировано ✨"), kb)
}

func (b *Bot) showMoveToSphere(params []string) error {
	itemPath, err := b.resolveLifeItemPath(params)
	if err != nil {
		return fmt.Errorf("move to sphere: %w", err)
	}

	if life.IsDocDir(itemPath) {
		if projectPath, ok := life.ProjectPathFromDoc(itemPath); ok {
			itemPath = projectPath
		}
	}

	spheres, err := life.ListSpheres(b.fs)
	if err != nil {
		return fmt.Errorf("move to sphere: %w", err)
	}

	var kb tg.Keyboard
	for _, spherePath := range spheres {
		moveParams := append(append([]string{}, params...), fs.ShortHash(spherePath))
		kb.AddRow(tg.NewBtn(
			life.SphereTitle(spherePath),
			tg.NewCmd(CmdMoveToSphere, moveParams),
		))
	}

	return b.showHTML(i18n.Tr("Переместить в сферу:"), &kb)
}

func (b *Bot) moveToSphere(params []string) error {
	if len(params) < 2 {
		return fmt.Errorf("move to sphere: missing params")
	}
	dstSphere, err := b.fs.ResolveDirParam(params[len(params)-1])
	if err != nil {
		return fmt.Errorf("move to sphere: %w", err)
	}
	itemPath, err := b.resolveLifeItemPath(params[:len(params)-1])
	if err != nil {
		return fmt.Errorf("move to sphere: %w", err)
	}

	if err := life.MoveEntry(b.fs, itemPath, dstSphere); err != nil {
		return fmt.Errorf("move to sphere: %w", err)
	}

	kb := tg.NewKeyboard([]tg.Row{
		tg.NewBtn(i18n.Tr("🌐 Сферы"), tg.NewCmd(CmdShowLifeSpheres, nil)),
		tg.NewBtn(i18n.Tr(i18n.StrHome), tg.NewCmd(CmdShowHome, nil)),
	})
	return b.showHTML(i18n.Tr("Перемещено ↔️"), kb)
}

func (b *Bot) resolveLifeItemPath(params []string) (string, error) {
	if len(params) >= 2 && params[1] != "" {
		dir, err := b.fs.ResolveDirParam(params[0])
		if err != nil {
			return "", err
		}
		filename, err := b.fs.Unhash(dir, params[1])
		if err != nil {
			return "", err
		}
		return fs.JoinDir(dir, filename), nil
	}
	return b.fs.ResolveDirParam(params[0])
}

func (b *Bot) setRecentLifeProject(projectPath string) {
	b.db.SetRecentCommand(CmdLifeRecentProject)
	b.db.SetRecentCommandParams([]string{fs.ShortHash(projectPath)})
}

func (b *Bot) recentLifeProject() string {
	cmd, ok := b.db.RecentCommand()
	if !ok || cmd != CmdLifeRecentProject {
		return ""
	}
	args, _ := b.db.RecentCommandParams()
	if len(args) == 0 {
		return ""
	}
	path, err := b.fs.ResolveDirParam(args[0])
	if err != nil {
		return ""
	}
	return path
}

func parseLifeShortcutTarget(text string) (target, rest string, ok bool) {
	parts := strings.SplitN(strings.TrimSpace(text), " ", 2)
	if len(parts) < 2 || !strings.Contains(parts[0], "/") {
		return "", text, false
	}
	return parts[0], strings.TrimSpace(parts[1]), true
}

func resolveLifeShortcutProject(fsys *fs.FS, shortcut string) (spherePath, areaPath string, err error) {
	shortcut = strings.Trim(shortcut, "/")
	parts := strings.Split(shortcut, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("shortcut project needs Sphere/Project format")
	}
	sphereName := parts[0]
	areaPath = life.SpherePath(sphereName)
	exists, err := fsys.Exists(areaPath, "")
	if err != nil {
		return "", "", err
	}
	if !exists {
		if err := life.Init(fsys); err != nil {
			return "", "", err
		}
	}
	for i, name := range parts[1:] {
		var child string
		var createErr error
		if i == 0 {
			child, createErr = life.CreateProject(fsys, areaPath, name)
		} else {
			child, createErr = life.CreateSection(fsys, areaPath, name)
		}
		if createErr != nil {
			return "", "", createErr
		}
		areaPath = child
	}
	return life.SpherePath(sphereName), areaPath, nil
}

func lifeKindLabel(kind life.Kind) string {
	switch kind {
	case life.KindDraft:
		return i18n.Tr("черновики")
	case life.KindFinal:
		return i18n.Tr("финальные")
	case life.KindDiscussion:
		return i18n.Tr("обсуждения")
	default:
		return ""
	}
}

func noteDetailKeyboard(dir, filename, dirHash string) *tg.Keyboard {
	row := tg.NewRow(
		tg.NewBtn("⬅️", noteBackCmd(dir)),
		tg.NewBtn("🔎", tg.NewCustomCmd(CmdInlineQuerySearchEveryWhere, nil, tg.CmdTypeInlineQueryCurrentChat)),
	)

	if life.IsDocDir(dir) {
		currentKind, _ := life.KindFromSubdir(baseName(dir))
		for _, kind := range []life.Kind{life.KindDraft, life.KindFinal, life.KindDiscussion} {
			if kind == currentKind {
				continue
			}
			row = append(row, tg.NewBtn(
				life.KindEmoji(kind),
				tg.NewCmd(CmdMoveNoteKind, []string{dirHash, fs.Hash(filename), life.KindCode(kind)}),
			))
		}
		row = append(row, tg.NewBtn("↔️", tg.NewCmd(CmdShowMoveNoteArea, []string{dirHash, fs.Hash(filename)})))
	}

	if canDeleteNote(dir, filename) {
		row = append(row, tg.NewBtn("🗑", tg.NewCmd(CmdShowDeleteFile, []string{dirHash, fs.Hash(filename)})))
	}
	row = append(row, tg.NewBtn("🏠", tg.NewCmd(CmdShowHome, nil)))
	return tg.NewKeyboard([]tg.Row{row})
}

func (b *Bot) moveNoteKind(params []string) error {
	if len(params) < 3 {
		return fmt.Errorf("move note kind: missing params")
	}
	dir, err := b.fs.ResolveDirParam(params[0])
	if err != nil {
		return fmt.Errorf("move note kind: %w", err)
	}
	filename, err := b.fs.Unhash(dir, params[1])
	if err != nil {
		return fmt.Errorf("move note kind: %w", err)
	}
	kind, ok := life.KindFromCode(params[2])
	if !ok {
		return fmt.Errorf("move note kind: bad kind")
	}
	if err := life.MoveDocKind(b.fs, dir, filename, kind); err != nil {
		return fmt.Errorf("move note kind: %w", err)
	}
	newDir := life.DocDir(lifeMustAreaPathFromDoc(dir), kind)
	return b.showFile([]string{fs.ShortHash(newDir), fs.Hash(filename)})
}

func (b *Bot) showMoveNoteArea(params []string) error {
	if len(params) < 2 {
		return fmt.Errorf("show move note area: missing params")
	}
	dirHash := params[0]
	filenameHash := params[1]
	srcDir, err := b.fs.ResolveDirParam(dirHash)
	if err != nil {
		return fmt.Errorf("show move note area: %w", err)
	}
	srcArea, ok := life.ProjectPathFromDoc(srcDir)
	if !ok {
		return fmt.Errorf("show move note area: not a life doc")
	}

	_ = life.EnsureSpheresRoot(b.fs)
	spheres, err := life.ListSpheres(b.fs)
	if err != nil {
		return fmt.Errorf("show move note area: %w", err)
	}

	var kb tg.Keyboard
	for _, spherePath := range spheres {
		areas, err := life.ListAllAreas(b.fs, spherePath)
		if err != nil {
			continue
		}
		for _, areaPath := range areas {
			if areaPath == srcArea {
				continue
			}
			kb.AddRow(tg.NewBtn(
				life.AreaPickerLabel(spherePath, areaPath),
				tg.NewCmd(CmdMoveNoteArea, []string{dirHash, filenameHash, fs.ShortHash(areaPath)}),
			))
		}
	}
	if len(kb.Btns) == 0 {
		kb.AddRow(tg.NewBtn("—", tg.NewCmd(CmdDoNothing, nil)))
	}
	kb.AddRow(tg.NewBtn("⬅️", tg.NewCmd(CmdShowFile, []string{dirHash, filenameHash})))
	return b.showHTML(i18n.Tr("Переместить заметку в область:"), &kb)
}

func (b *Bot) moveNoteArea(params []string) error {
	if len(params) < 3 {
		return fmt.Errorf("move note area: missing params")
	}
	dir, err := b.fs.ResolveDirParam(params[0])
	if err != nil {
		return fmt.Errorf("move note area: %w", err)
	}
	filename, err := b.fs.Unhash(dir, params[1])
	if err != nil {
		return fmt.Errorf("move note area: %w", err)
	}
	dstArea, err := b.fs.ResolveDirParam(params[2])
	if err != nil {
		return fmt.Errorf("move note area: %w", err)
	}
	if err := life.MoveDocToArea(b.fs, dir, filename, dstArea); err != nil {
		return fmt.Errorf("move note area: %w", err)
	}
	kind, _ := life.KindFromSubdir(baseName(dir))
	newDir := life.DocDir(dstArea, kind)
	return b.showFile([]string{fs.ShortHash(newDir), fs.Hash(filename)})
}

func lifeMustAreaPathFromDoc(docDir string) string {
	area, ok := life.ProjectPathFromDoc(docDir)
	if !ok {
		return ""
	}
	return area
}

func noteBackCmd(dir string) tg.Cmd {
	if projectPath, ok := life.ProjectPathFromDoc(dir); ok {
		if kind, ok := life.KindFromSubdir(baseName(dir)); ok {
			return tg.NewCmd(CmdShowLifeDocs, []string{fs.ShortHash(projectPath), life.KindCode(kind)})
		}
		return tg.NewCmd(CmdShowLifeProject, []string{fs.ShortHash(projectPath)})
	}
	return tg.NewCmd(CmdShowFiles, nil)
}

func lifeFinalizeBtn(dir, filename string) *tg.Btn {
	if baseName(dir) != life.SubDirDrafts {
		return nil
	}
	btn := tg.NewBtn("✨", tg.NewCmd(CmdFinalizeDoc, []string{fs.ShortHash(dir), fs.ShortHash(filename)}))
	return &btn
}

func lifeMoveSphereBtn(dir, filename string) *tg.Btn {
	var params []string
	if filename == "" {
		params = []string{fs.ShortHash(dir)}
	} else {
		params = []string{fs.ShortHash(dir), fs.ShortHash(filename)}
	}
	itemPath := dir
	if filename != "" {
		itemPath = fs.JoinDir(dir, filename)
	}
	if !life.IsLifePath(itemPath) {
		return nil
	}
	btn := tg.NewBtn("↔️", tg.NewCmd(CmdShowMoveToSphere, params))
	return &btn
}

func baseName(path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}
