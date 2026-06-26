package stt

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTranscribeGroq(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/openai/v1/audio/transcriptions" {
			t.Fatalf("path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("auth: %q", got)
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("model") != groqWhisperModel {
			t.Fatalf("model: %q", r.FormValue("model"))
		}
		if r.FormValue("language") != "ru" {
			t.Fatalf("language: %q", r.FormValue("language"))
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		if !strings.HasSuffix(header.Filename, ".ogg") {
			t.Fatalf("filename: %q", header.Filename)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"text": "Привет мир"})
	}))
	defer srv.Close()

	oldURL := groqTranscribeURL
	groqTranscribeURL = srv.URL + "/openai/v1/audio/transcriptions"
	defer func() { groqTranscribeURL = oldURL }()

	text, err := transcribeGroq("test-key", []byte("fake-audio"), "audio/ogg", "ru")
	if err != nil {
		t.Fatal(err)
	}
	if text != "Привет мир" {
		t.Fatalf("got %q", text)
	}
}

func TestTranscribe_PrefersGroq(t *testing.T) {
	var called string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = "groq"
		_ = json.NewEncoder(w).Encode(map[string]string{"text": "ok"})
	}))
	defer srv.Close()

	oldURL := groqTranscribeURL
	groqTranscribeURL = srv.URL
	defer func() { groqTranscribeURL = oldURL }()

	text, err := Transcribe("groq-key", "kie-key", []byte("x"), "audio/ogg", "ru")
	if err != nil {
		t.Fatal(err)
	}
	if called != "groq" {
		t.Fatalf("expected groq, got %q", called)
	}
	if text != "ok" {
		t.Fatalf("got %q", text)
	}
}

func TestGroqLanguage(t *testing.T) {
	if got := groqLanguage("ru"); got != "ru" {
		t.Fatalf("got %q", got)
	}
	if got := groqLanguage("en-US"); got != "en" {
		t.Fatalf("got %q", got)
	}
	if got := groqLanguage(""); got != "" {
		t.Fatalf("got %q", got)
	}
}
