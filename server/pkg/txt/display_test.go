package txt

import "testing"

func TestDisplayText_PrefersTranscriptOverVoicePlaceholder(t *testing.T) {
	raw := "Расписание на завтра\n\n![](media/tg_voice.ogg)"
	if got := DisplayText(raw); got != "Расписание на завтра" {
		t.Fatalf("got %q", got)
	}
}

func TestDisplayText_FallsBackToPlaceholder(t *testing.T) {
	raw := VoicePlaceholder + "\n\n![](media/tg_voice.ogg)"
	if got := DisplayText(raw); got != VoicePlaceholder {
		t.Fatalf("got %q", got)
	}
}
