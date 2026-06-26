// Command sttprobe checks voice STT configuration (Groq preferred, kie.ai fallback).
//
// Usage (inside Docker with .env mounted):
//
//	/app/sttprobe
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/zakirullin/files.md/server/config"
	"github.com/zakirullin/files.md/server/stt"
)

const publicMP3 = "https://file.aiquickdraw.com/custom-page/akr/section-images/1757157053357tn37vxc8.mp3"

func main() {
	config.LoadDotEnv()
	if err := config.LoadBotConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	groqKey := config.ServerCfg.GroqAPIKey
	kieKey := config.ServerCfg.KieAPIKey

	switch {
	case groqKey != "":
		fmt.Printf("groq api key loaded (%d chars)\n", len(groqKey))
	case kieKey != "":
		fmt.Printf("kie api key loaded (%d chars)\n", len(kieKey))
	default:
		fmt.Fprintln(os.Stderr, "set GROQ_API_KEY or KIE_API_KEY in .env")
		os.Exit(1)
	}

	audio, err := downloadPublicMP3()
	if err != nil {
		fmt.Fprintf(os.Stderr, "download test mp3: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("test audio: %d bytes\n", len(audio))

	text, err := stt.Transcribe(groqKey, kieKey, audio, "audio/mpeg", "ru")
	if err != nil {
		fmt.Fprintf(os.Stderr, "STT failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("STT ok: %q\n", text)
}

func downloadPublicMP3() ([]byte, error) {
	resp, err := http.Get(publicMP3)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	if len(body) < 1000 {
		return nil, fmt.Errorf("too small (%d bytes)", len(body))
	}
	return body, nil
}
