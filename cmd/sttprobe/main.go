// Command sttprobe runs the kie.ai upload → createTask → poll chain for debugging.
//
// Usage (inside Docker with .env mounted):
//
//	go run ./cmd/sttprobe
package main

import (
	"fmt"
	"os"

	"github.com/zakirullin/files.md/server/config"
	"github.com/zakirullin/files.md/server/stt"
)

func main() {
	config.LoadDotEnv()
	if err := config.LoadBotConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	apiKey := config.ServerCfg.KieAPIKey
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "KIE_API_KEY is not set")
		os.Exit(1)
	}

	audio := sampleOGG()
	fmt.Printf("probing STT with %d bytes of sample ogg audio...\n", len(audio))

	text, err := stt.Transcribe(apiKey, audio, "audio/ogg", "ru")
	if err != nil {
		fmt.Fprintf(os.Stderr, "STT failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("STT ok: %q\n", text)
}

// sampleOGG is a tiny valid OGG container (silence). Enough for upload/task wiring tests.
func sampleOGG() []byte {
	return []byte{
		0x4f, 0x67, 0x67, 0x53, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x1e,
		0x01, 0x76, 0x6f, 0x72, 0x62, 0x69, 0x73, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}
