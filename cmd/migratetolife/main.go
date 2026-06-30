// One-time migration from legacy note folders into the life/projects structure.
//
// Usage:
//
//	go run ./cmd/migratetolife --dst ./storage/USER_ID
//	go run ./cmd/migratetolife --dst ./storage/USER_ID --dirs notes,brain --sphere Личное --kind final
//	go run ./cmd/migratetolife --dst ./storage/USER_ID --root-md --dry-run
//	go run ./cmd/migratetolife --dst ./storage/USER_ID --per-dir-sphere
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/afero"

	"github.com/zakirullin/files.md/server/fs"
	"github.com/zakirullin/files.md/server/life"
)

func main() {
	dst := flag.String("dst", "", "Files.md user storage directory (required)")
	dirs := flag.String("dirs", "", "Comma-separated legacy folders to migrate (default: all note folders)")
	sphere := flag.String("sphere", "", "Target sphere for all folders (default: Личное)")
	kind := flag.String("kind", "final", "Document kind: draft, final, discussion")
	rootMD := flag.Bool("root-md", false, "Also migrate root-level .md files into a new project")
	rootProject := flag.String("root-project", "Inbox", "Project name for root-level notes")
	copyOnly := flag.Bool("copy", false, "Copy files instead of moving")
	dryRun := flag.Bool("dry-run", false, "Print planned actions without writing")
	perDirSphere := flag.Bool("per-dir-sphere", false, "Pick sphere from each folder name")
	flag.Parse()

	if *dst == "" {
		fmt.Fprintln(os.Stderr, "--dst is required")
		os.Exit(1)
	}

	k, ok := parseKind(*kind)
	if !ok {
		fmt.Fprintln(os.Stderr, "--kind must be draft, final, or discussion")
		os.Exit(1)
	}

	userFS, err := fs.NewFS(*dst, afero.NewOsFs())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	selected := splitCSV(*dirs)

	if *perDirSphere && *sphere != "" {
		fmt.Fprintln(os.Stderr, "use either --sphere or --per-dir-sphere")
		os.Exit(1)
	}

	var total life.MigrateResult
	if *perDirSphere {
		names, err := life.LegacyNoteDirNames(userFS, selected)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, dirName := range names {
			res, err := life.MigrateFromFolders(userFS, life.MigrateOptions{
				Sphere: life.SphereForLegacyDir(dirName),
				Kind:   k,
				Dirs:   []string{dirName},
				Copy:   *copyOnly,
				DryRun: *dryRun,
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			mergeResult(&total, &res)
		}
	} else {
		targetSphere := *sphere
		if targetSphere == "" {
			targetSphere = "Личное"
		}
		res, err := life.MigrateFromFolders(userFS, life.MigrateOptions{
			Sphere:      targetSphere,
			Kind:        k,
			Dirs:        selected,
			RootMD:      *rootMD,
			RootProject: *rootProject,
			Copy:        *copyOnly,
			DryRun:      *dryRun,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		total = res
	}

	if *perDirSphere && *rootMD {
		res, err := life.MigrateFromFolders(userFS, life.MigrateOptions{
			Sphere:      "Личное",
			Kind:        k,
			RootMD:      true,
			RootProject: *rootProject,
			Copy:        *copyOnly,
			DryRun:      *dryRun,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		mergeResult(&total, &res)
	}

	printResult(total, *dryRun)
}

func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func parseKind(s string) (life.Kind, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "draft", "drafts", "d", "черновик", "черновики":
		return life.KindDraft, true
	case "final", "finals", "f", "финал", "финальные":
		return life.KindFinal, true
	case "discussion", "discussions", "c", "обсуждение", "обсуждения":
		return life.KindDiscussion, true
	default:
		return life.KindFinal, false
	}
}

func printResult(res life.MigrateResult, dryRun bool) {
	prefix := "Done"
	if dryRun {
		prefix = "Dry run"
	}
	fmt.Printf("%s: %d folders, %d projects, %d sections, %d files moved, %d skipped\n",
		prefix, res.DirsProcessed, res.ProjectsCreated, res.SectionsCreated, res.FilesMoved, res.FilesSkipped)
}

func mergeResult(total, part *life.MigrateResult) {
	total.DirsProcessed += part.DirsProcessed
	total.ProjectsCreated += part.ProjectsCreated
	total.SectionsCreated += part.SectionsCreated
	total.FilesMoved += part.FilesMoved
	total.FilesSkipped += part.FilesSkipped
}
