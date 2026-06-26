package stt

import "fmt"

// Transcribe returns speech-to-text using the first configured provider:
// Groq (preferred), then kie.ai.
func Transcribe(groqKey, kieKey string, audio []byte, mimeType, languageCode string) (string, error) {
	if groqKey != "" {
		return transcribeGroq(groqKey, audio, mimeType, languageCode)
	}
	if kieKey != "" {
		return transcribeKie(kieKey, audio, mimeType, languageCode)
	}
	return "", fmt.Errorf("stt: set GROQ_API_KEY or KIE_API_KEY")
}

// Enabled reports whether any STT provider is configured.
func Enabled(groqKey, kieKey string) bool {
	return groqKey != "" || kieKey != ""
}
