// Package stt transcribes audio via kie.ai (ElevenLabs speech-to-text).
package stt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	uploadURL     = "https://kieai.redpandaai.co/api/file-base64-upload"
	createTaskURL = "https://api.kie.ai/api/v1/jobs/createTask"
	recordInfoURL = "https://api.kie.ai/api/v1/jobs/recordInfo"
	model         = "elevenlabs/speech-to-text"
)

// Transcribe uploads audio to kie.ai and returns the transcript text.
func Transcribe(apiKey string, audio []byte, mimeType string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("stt: KIE_API_KEY is not set")
	}
	if len(audio) == 0 {
		return "", fmt.Errorf("stt: empty audio")
	}
	if mimeType == "" {
		mimeType = "audio/ogg"
	}

	downloadURL, err := uploadAudio(apiKey, audio, mimeType)
	if err != nil {
		return "", err
	}

	taskID, err := createTask(apiKey, downloadURL)
	if err != nil {
		return "", err
	}

	return pollTask(apiKey, taskID)
}

func uploadAudio(apiKey string, audio []byte, mimeType string) (string, error) {
	ext := ".ogg"
	switch mimeType {
	case "audio/mpeg", "audio/mp3":
		ext = ".mp3"
	case "audio/wav", "audio/x-wav":
		ext = ".wav"
	case "audio/mp4":
		ext = ".m4a"
	}

	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(audio))
	body, _ := json.Marshal(map[string]string{
		"base64Data": dataURL,
		"uploadPath": "filesmd/voice",
		"fileName":   fmt.Sprintf("voice%s", ext),
	})

	req, err := http.NewRequest(http.MethodPost, uploadURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("stt upload: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var parsed struct {
		Success bool `json:"success"`
		Code    int  `json:"code"`
		Msg     string
		Data    struct {
			DownloadURL string `json:"downloadUrl"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("stt upload: parse response: %w", err)
	}
	if parsed.Data.DownloadURL == "" {
		return "", fmt.Errorf("stt upload failed: %s", strings.TrimSpace(parsed.Msg))
	}
	return parsed.Data.DownloadURL, nil
}

func createTask(apiKey, audioURL string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model": model,
		"input": map[string]any{
			"audio_url":        audioURL,
			"language_code":    "ru",
			"tag_audio_events": false,
			"diarize":          false,
		},
	})

	req, err := http.NewRequest(http.MethodPost, createTaskURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("stt create task: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var parsed struct {
		Code int
		Msg  string
		Data struct {
			TaskID string `json:"taskId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("stt create task: parse response: %w", err)
	}
	if parsed.Data.TaskID == "" {
		return "", fmt.Errorf("stt create task failed: %s", strings.TrimSpace(parsed.Msg))
	}
	return parsed.Data.TaskID, nil
}

func pollTask(apiKey, taskID string) (string, error) {
	deadline := time.Now().Add(3 * time.Minute)
	interval := 2 * time.Second

	for time.Now().Before(deadline) {
		text, state, err := taskStatus(apiKey, taskID)
		if err != nil {
			return "", err
		}
		switch state {
		case "success":
			if text == "" {
				return "", fmt.Errorf("stt: empty transcript")
			}
			return text, nil
		case "fail":
			return "", fmt.Errorf("stt: transcription failed")
		}
		time.Sleep(interval)
		if interval < 8*time.Second {
			interval += time.Second
		}
	}
	return "", fmt.Errorf("stt: timeout waiting for task %s", taskID)
}

func taskStatus(apiKey, taskID string) (text, state string, err error) {
	req, err := http.NewRequest(http.MethodGet, recordInfoURL+"?taskId="+taskID, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("stt poll: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var parsed struct {
		Data struct {
			State      string `json:"state"`
			ResultJSON string `json:"resultJson"`
			FailMsg    string `json:"failMsg"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", "", fmt.Errorf("stt poll: parse response: %w", err)
	}
	if parsed.Data.State == "fail" {
		return "", "fail", fmt.Errorf("stt: %s", parsed.Data.FailMsg)
	}
	if parsed.Data.State != "success" {
		return "", parsed.Data.State, nil
	}
	return extractTranscript(parsed.Data.ResultJSON), "success", nil
}

func extractTranscript(resultJSON string) string {
	if resultJSON == "" {
		return ""
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(resultJSON), &raw); err != nil {
		return strings.TrimSpace(resultJSON)
	}
	for _, key := range []string{"text", "transcript", "transcription"} {
		if v, ok := raw[key].(string); ok && v != "" {
			return strings.TrimSpace(v)
		}
	}
	if obj, ok := raw["resultObject"].(map[string]any); ok {
		for _, key := range []string{"text", "transcript", "transcription"} {
			if v, ok := obj[key].(string); ok && v != "" {
				return strings.TrimSpace(v)
			}
		}
	}
	if urls, ok := raw["resultUrls"].([]any); ok && len(urls) > 0 {
		if u, ok := urls[0].(string); ok {
			return fetchTextURL(u)
		}
	}
	return ""
}

func fetchTextURL(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(body))
}
