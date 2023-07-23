package journal

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Kunde21/markdownfmt/v3/markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"zakirullin/stuffbot/internal/fs"
	pkgText "zakirullin/stuffbot/pkg/text"
)

var now = time.Now // to be replaced in tests

var re = regexp.MustCompile(`\n+`)

const (
	headerLevel = 4
)

func AddDailyNote(dir, noteFilename string, botFs *fs.FS, journalFilenameFormat, journalHeaderFormat string) error {
	noteContent, err := botFs.Content(dir, noteFilename)
	if err != nil {
		return fmt.Errorf("failed to move to journal: can't get note content: %w", err)
	}
	note := fs.Title(noteFilename)
	noteContent = strings.TrimSpace(noteContent)
	if noteContent != "" {
		note += "\n" + noteContent
	}
	// Replace all occurrence of one or several multiples in a row with exactly two newlines, to comply with markdown
	note = re.ReplaceAllString(pkgText.NormNewLines(note), "\n\n")

	journalFilename := now().Format(journalFilenameFormat)
	exists, err := botFs.Exists(fs.DirJournal, journalFilename)
	if err != nil {
		return err
	}
	var md string
	if exists {
		md, err = botFs.Content(fs.DirJournal, journalFilename)
		if err != nil {
			return err
		}
		md = pkgText.NormNewLines(md)
	}

	md = insertDailyNote(md, journalHeaderFormat, note)
	return botFs.Put(fs.DirJournal, journalFilename, md)
}

func insertDailyNote(mdContent, journalHeaderFormat, note string) string {
	header := now().Format(journalHeaderFormat)
	r := markdown.NewRenderer()
	md := goldmark.New(
		goldmark.WithRenderer(r),
	)

	var buf bytes.Buffer

	source := []byte(mdContent)
	root := md.Parser().Parse(text.NewReader(source))
	addJournalRecordAfterHeader(source, root, header, note)

	err := r.Render(&buf, source, root)
	if err != nil {
		panic(err) // should never happen
	}
	return buf.String()
}

func addJournalRecordAfterHeader(source []byte, root ast.Node, headerText, txt string) {
	listItem := ast.NewListItem(0)
	listItem.AppendChild(listItem, ast.NewString([]byte(txt)))
	var header ast.Node
	var noteInserted bool

	walker := func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if header != nil {
			// we have already found the header, so we are looking for the end of the section:
			// next header with the same or higher level to insert the note before it
			if h, ok := node.(*ast.Heading); ok && entering && h.Level <= headerLevel {
				if h.PreviousSibling() != header {
					// If the note doesn't go right after the corresponding header, so we need to insert a separator
					h.InsertBefore(root, h, newSeparator())
				}

				h.InsertBefore(root, h, newJournalRecord(txt))
				noteInserted = true
				return ast.WalkStop, nil
			}
			return ast.WalkContinue, nil
		}

		if h, ok := node.(*ast.Heading); ok && entering {
			// if it is a header, let's check if it is the header we are looking for
			if string(h.Text(source)) == headerText && h.Level == headerLevel {
				header = h
			}
		}

		return ast.WalkContinue, nil
	}
	err := ast.Walk(root, walker)
	if err != nil {
		// walker() doesn't return errors, so err must always be nil
		panic(err)
	}
	if !noteInserted { // Insert the note at the end of the document
		if header == nil {
			header = newHeader(headerText)
			root.AppendChild(root, header)
		}
		if root.LastChild() != header {
			// If the note doesn't go right after the corresponding header, so we need to insert a separator
			root.AppendChild(root, newSeparator())
		}
		root.AppendChild(root, newJournalRecord(txt))
	}
}

func newHeader(header string) *ast.Heading {
	heading := ast.NewHeading(headerLevel)
	heading.AppendChild(heading, ast.NewString([]byte(header)))
	return heading
}

func newJournalRecord(txt string) ast.Node {
	record := ast.NewParagraph()
	record.AppendChild(record, ast.NewString([]byte(txt)))
	return record
}

func newSeparator() ast.Node {
	return ast.NewThematicBreak()
}
