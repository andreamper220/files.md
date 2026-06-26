package txt

import (
	"strings"
	"testing"
)

func TestIsVoiceDraft(t *testing.T) {
	raw := "Короткое саммари.\nПолная расшифровка.\n\n![](media/tg_voice.oga)"
	if !IsVoiceDraft(raw) {
		t.Fatal("expected voice draft")
	}
	if IsVoiceDraft("Hello\n\n![](media/photo.jpg)") {
		t.Fatal("photo is not voice")
	}
}

func TestVoiceTitleSuggestions(t *testing.T) {
	raw := "Короткое саммари.\nПолная расшифровка голосового сообщения.\n\n![](media/tg_voice.oga)"
	got := VoiceTitleSuggestions(raw)
	if len(got) == 0 {
		t.Fatal("expected suggestions")
	}
	if got[0] != "Короткое саммари." {
		t.Fatalf("first %q", got[0])
	}
}

func TestApplyVoiceDraftTitle(t *testing.T) {
	raw := "Старое саммари.\nПолная расшифровка голосового.\n\n![](media/tg_voice.oga)"
	got := ApplyVoiceDraftTitle(raw, "Мой заголовок")
	if DraftTitle(got) != "Мой заголовок" {
		t.Fatalf("title %q", DraftTitle(got))
	}
	if !strings.Contains(got, "Полная расшифровка голосового.") {
		t.Fatalf("missing transcript: %q", got)
	}
	if !strings.Contains(got, "![](media/tg_voice.oga)") {
		t.Fatalf("missing media: %q", got)
	}
}
