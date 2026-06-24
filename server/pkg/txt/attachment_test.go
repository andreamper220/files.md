package txt

import (
	"strings"
	"testing"
)

func TestNeedsUserTitle_AttachmentOnly(t *testing.T) {
	raw := FormatAttachmentContent("media/tg_abc.pdf")
	if !NeedsUserTitle(raw) {
		t.Fatal("expected attachment-only draft to need title")
	}
}

func TestNeedsUserTitle_WithCaption(t *testing.T) {
	raw := "My note\n" + FormatAttachmentContent("media/tg_abc.pdf")
	if NeedsUserTitle(raw) {
		t.Fatal("caption should count as title")
	}
}

func TestApplyDraftTitle_LabelsAttachment(t *testing.T) {
	raw := FormatAttachmentContent("media/tg_abc.pdf")
	got := ApplyDraftTitle(raw, "Отчёт")
	if !strings.Contains(got, "📎 [Отчёт](media/tg_abc.pdf)") {
		t.Fatalf("got %q", got)
	}
	if DraftTitle(got) != "Отчёт" {
		t.Fatalf("title %q", DraftTitle(got))
	}
}
