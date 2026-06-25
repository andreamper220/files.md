package stt

import "testing"

func TestExtractTranscript_ResultObject(t *testing.T) {
	raw := `{"resultObject":{"text":"Привет мир","language_code":"ru"}}`
	if got := extractTranscript(raw); got != "Привет мир" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractTranscript_DirectText(t *testing.T) {
	raw := `{"text":"Расписание на завтра"}`
	if got := extractTranscript(raw); got != "Расписание на завтра" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractTranscript_TranscriptsArray(t *testing.T) {
	raw := `{"transcripts":[{"text":"Первая фраза"},{"text":"Вторая фраза"}]}`
	want := "Первая фраза\nВторая фраза"
	if got := extractTranscript(raw); got != want {
		t.Fatalf("got %q", got)
	}
}

func TestAudioExt(t *testing.T) {
	if got := audioExt("audio/ogg"); got != ".ogg" {
		t.Fatalf("got %q", got)
	}
	if got := audioExt("audio/mpeg"); got != ".mp3" {
		t.Fatalf("got %q", got)
	}
}
