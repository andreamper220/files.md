// Groq Whisper: https://console.groq.com/docs/speech-to-text
package stt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

var groqTranscribeURL = "https://api.groq.com/openai/v1/audio/transcriptions"

const groqWhisperModel = "whisper-large-v3-turbo"

func transcribeGroq(apiKey string, audio []byte, mimeType, languageCode string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("stt: GROQ_API_KEY is not set")
	}
	if len(audio) == 0 {
		return "", fmt.Errorf("stt: empty audio")
	}
	if mimeType == "" {
		mimeType = "audio/ogg"
	}

	body, contentType, err := groqMultipartBody(audio, mimeType, languageCode)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, groqTranscribeURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("stt groq: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("stt groq: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		msg := strings.TrimSpace(string(raw))
		if msg == "" {
			msg = resp.Status
		}
		return "", fmt.Errorf("stt groq failed (http %d): %s", resp.StatusCode, msg)
	}

	var parsed struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("stt groq: parse response: %w", err)
	}
	text := strings.TrimSpace(parsed.Text)
	if text == "" {
		return "", fmt.Errorf("stt groq: empty transcript")
	}
	return text, nil
}

func groqMultipartBody(audio []byte, mimeType, languageCode string) (io.Reader, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fileName := "voice" + audioExt(mimeType)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, "", fmt.Errorf("stt groq: %w", err)
	}
	if _, err := part.Write(audio); err != nil {
		return nil, "", fmt.Errorf("stt groq: %w", err)
	}
	if err := writer.WriteField("model", groqWhisperModel); err != nil {
		return nil, "", fmt.Errorf("stt groq: %w", err)
	}
	if err := writer.WriteField("response_format", "json"); err != nil {
		return nil, "", fmt.Errorf("stt groq: %w", err)
	}
	if lang := groqLanguage(languageCode); lang != "" {
		if err := writer.WriteField("language", lang); err != nil {
			return nil, "", fmt.Errorf("stt groq: %w", err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("stt groq: %w", err)
	}
	return &buf, writer.FormDataContentType(), nil
}

func groqLanguage(languageCode string) string {
	languageCode = strings.TrimSpace(strings.ToLower(languageCode))
	switch languageCode {
	case "", "auto":
		return ""
	default:
		// ISO-639-1: ru, en, ...
		if idx := strings.IndexByte(languageCode, '-'); idx > 0 {
			languageCode = languageCode[:idx]
		}
		if len(languageCode) == 2 {
			return languageCode
		}
		return ""
	}
}
