package txt

import (
	"strings"
	"testing"
)

func TestNeedsUserTitle_AttachmentOnly(t *testing.T) {
	raw := FormatAttachmentContent("media/tg_abc.pdf", "")
	if !NeedsUserTitle(raw) {
		t.Fatal("expected unnamed attachment-only draft to need title")
	}
}

func TestNeedsUserTitle_NamedAttachmentOnly(t *testing.T) {
	raw := FormatAttachmentContent("media/go-cheatsheet.pdf", "go-cheatsheet.pdf")
	if NeedsUserTitle(raw) {
		t.Fatal("expected named attachment-only draft to skip title prompt")
	}
}

func TestNeedsUserTitle_WithCaption(t *testing.T) {
	raw := "My note\n" + FormatAttachmentContent("media/tg_abc.pdf", "report.pdf")
	if NeedsUserTitle(raw) {
		t.Fatal("caption should count as title")
	}
}

func TestApplyDraftTitle_LabelsAttachment(t *testing.T) {
	raw := FormatAttachmentContent("media/tg_abc.pdf", "")
	got := ApplyDraftTitle(raw, "Отчёт")
	if !strings.Contains(got, "📎 [Отчёт](media/tg_abc.pdf)") {
		t.Fatalf("got %q", got)
	}
	if DraftTitle(got) != "Отчёт" {
		t.Fatalf("title %q", DraftTitle(got))
	}
}

func TestApplyDraftTitle_KeepsNamedAttachments(t *testing.T) {
	raw := strings.Join([]string{
		FormatAttachmentContent("media/a.pdf", "a.pdf"),
		FormatAttachmentContent("media/b.pdf", "b.pdf"),
	}, "\n")
	got := ApplyDraftTitle(raw, "Общий заголовок")
	if !strings.Contains(got, "📎 [a.pdf](media/a.pdf)") {
		t.Fatalf("first file renamed: %q", got)
	}
	if !strings.Contains(got, "📎 [b.pdf](media/b.pdf)") {
		t.Fatalf("second file renamed: %q", got)
	}
}

func TestAttachmentNoteTitle(t *testing.T) {
	got := AttachmentNoteTitle([]string{"a.pdf", "b.pdf"})
	if got != "a.pdf + b.pdf" {
		t.Fatalf("got %q", got)
	}
}

func TestParseAttachments_Multiple(t *testing.T) {
	raw := strings.Join([]string{
		FormatAttachmentContent("media/a.pdf", "a.pdf"),
		FormatAttachmentContent("media/b.pdf", "b.pdf"),
	}, "\n")
	got := ParseAttachments(raw)
	if len(got) != 2 {
		t.Fatalf("got %d attachments", len(got))
	}
}
