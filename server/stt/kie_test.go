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

func TestOrderURLsForSTT_PrefersDownload(t *testing.T) {
	urls := orderURLsForSTT([]string{
		"https://tempfile.redpandaai.co/kieai/1/voice.ogg",
		"https://kieai.redpandaai.co/download/file_abc",
		"https://kieai.redpandaai.co/files/voice.ogg",
	})
	if urls[0] != "https://kieai.redpandaai.co/download/file_abc" {
		t.Fatalf("got %v", urls)
	}
}

func TestIsTempfileURL(t *testing.T) {
	if !isTempfileURL("https://tempfile.redpandaai.co/x.ogg") {
		t.Fatal("expected tempfile url")
	}
	if isTempfileURL("https://kieai.redpandaai.co/files/x.ogg") {
		t.Fatal("expected non-tempfile url")
	}
}

func TestUniqueURLs_PrefersFileURLFirst(t *testing.T) {
	got := uniqueURLs("https://example.com/file.ogg", "https://example.com/download/abc")
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
	if got[0] != "https://example.com/file.ogg" {
		t.Fatalf("got %q", got[0])
	}
}

func TestSttInputVariants_MinimalFirst(t *testing.T) {
	variants := sttInputVariants("https://example.com/voice.ogg", "ru")
	if len(variants) < 1 {
		t.Fatalf("expected variants")
	}
	if len(variants[0]) != 1 {
		t.Fatalf("first variant should be minimal, got %v", variants[0])
	}
}
