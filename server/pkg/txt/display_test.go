package txt

import (
	"strings"
	"testing"
)

func TestDisplayText_PrefersTranscriptOverVoicePlaceholder(t *testing.T) {
	raw := "Расписание на завтра\n\n![](media/tg_voice.ogg)"
	if got := DisplayText(raw); got != "Расписание на завтра" {
		t.Fatalf("got %q", got)
	}
}

func TestDisplayText_UsesVoiceSummary(t *testing.T) {
	raw := "Короткое саммари.\nПолная расшифровка голосового сообщения с деталями.\n\n![](media/tg_voice.ogg)"
	if got := DisplayText(raw); got != "Короткое саммари." {
		t.Fatalf("got %q", got)
	}
}

func TestVoiceDetailBody_ShowsSummaryAndTranscript(t *testing.T) {
	raw := "Короткое саммари.\nПолная расшифровка.\n\n![](media/tg_voice.ogg)"
	got := VoiceDetailBody(raw)
	want := "Короткое саммари.\n\nПолная расшифровка.\n\n![](media/tg_voice.ogg)"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestDisplayText_FallsBackToPlaceholder(t *testing.T) {
	raw := VoicePlaceholder + "\n\n![](media/tg_voice.ogg)"
	if got := DisplayText(raw); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestVoiceSummary_AfterPlaceholder(t *testing.T) {
	raw := VoicePlaceholder + "\nРасписание на завтра.\n\n![](media/tg_voice.ogg)"
	if got := VoiceSummary(raw); got != "Расписание на завтра." {
		t.Fatalf("got %q", got)
	}
}

func TestFormatNoteDetailBody_MixedVoice(t *testing.T) {
	raw := "тест\n" + VoicePlaceholder + "\nРасшифровка голоса.\n\n![](media/tg_voice.ogg)"
	got := FormatNoteDetailBody(raw)
	if !strings.Contains(got, "тест") {
		t.Fatalf("missing text: %q", got)
	}
	if strings.Contains(got, VoicePlaceholder) {
		t.Fatalf("placeholder still visible: %q", got)
	}
	if !strings.Contains(got, "Расшифровка голоса.") {
		t.Fatalf("missing transcript: %q", got)
	}
}

func TestSummarizeVoice_TruncatesLongText(t *testing.T) {
	long := stringsRepeat("слово ", 30)
	got := SummarizeVoice(long)
	if utf8Count(got) > voiceSummaryMaxRunes {
		t.Fatalf("summary too long: %d runes", utf8Count(got))
	}
}

func stringsRepeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func utf8Count(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}
