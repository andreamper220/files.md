// Package stt transcribes audio via kie.ai (ElevenLabs speech-to-text).
// API docs: https://docs.kie.ai/market/quickstart
package stt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const (
	// File uploads: https://docs.kie.ai/file-upload-api/quickstart
	streamUploadURL = "https://kieai.redpandaai.co/api/file-stream-upload"
	uploadPath      = "filesmd/voice"

	// Market jobs: https://docs.kie.ai/market/quickstart
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

func audioExt(mimeType string) string {
	switch mimeType {
	case "audio/mpeg", "audio/mp3":
		return ".mp3"
	case "audio/wav", "audio/x-wav":
		return ".wav"
	case "audio/mp4", "audio/aac":
		return ".m4a"
	default:
		return ".ogg"
	}
}

func uploadAudio(apiKey string, audio []byte, mimeType string) (string, error) {
	fileName := fmt.Sprintf("voice_%d%s", time.Now().UnixNano(), audioExt(mimeType))

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", fmt.Errorf("stt upload: %w", err)
	}
	if _, err := part.Write(audio); err != nil {
		return "", fmt.Errorf("stt upload: %w", err)
	}
	if err := writer.WriteField("uploadPath", uploadPath); err != nil {
		return "", fmt.Errorf("stt upload: %w", err)
	}
	if err := writer.WriteField("fileName", fileName); err != nil {
		return "", fmt.Errorf("stt upload: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("stt upload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, streamUploadURL, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

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
			FileURL     string `json:"fileUrl"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("stt upload: parse response: %w", err)
	}
	audioURL := strings.TrimSpace(parsed.Data.DownloadURL)
	if audioURL == "" {
		audioURL = strings.TrimSpace(parsed.Data.FileURL)
	}
	if parsed.Code != 200 || !parsed.Success || audioURL == "" {
		msg := strings.TrimSpace(parsed.Msg)
		if msg == "" {
			msg = "no download URL in response"
		}
		if parsed.Code != 0 && parsed.Code != 200 {
			return "", fmt.Errorf("stt upload failed (code %d): %s", parsed.Code, msg)
		}
		return "", fmt.Errorf("stt upload failed: %s", msg)
	}
	return audioURL, nil
}

func createTask(apiKey, audioURL string) (string, error) {
	// Params per https://kieai.mintlify.app/market/elevenlabs/speech-to-text
	body, _ := json.Marshal(map[string]any{
		"model": model,
		"input": map[string]any{
			"audio_url":        audioURL,
			"language_code":    "",
			"tag_audio_events": true,
			"diarize":          false,
		},
	})

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		taskID, err := createTaskOnce(apiKey, body)
		if err == nil {
			return taskID, nil
		}
		lastErr = err
		if attempt < 3 && isRetryableSTTError(err) {
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
		return "", err
	}
	return "", lastErr
}

func isRetryableSTTError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "server exception") ||
		strings.Contains(msg, "try again") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "code 500") ||
		strings.Contains(msg, "code 455")
}

func createTaskOnce(apiKey string, body []byte) (string, error) {
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
	if parsed.Code == 200 && parsed.Data.TaskID != "" {
		return parsed.Data.TaskID, nil
	}
	msg := strings.TrimSpace(parsed.Msg)
	if msg == "" {
		msg = "empty taskId"
	}
	if parsed.Code != 0 {
		return "", fmt.Errorf("stt create task failed (code %d): %s", parsed.Code, msg)
	}
	return "", fmt.Errorf("stt create task failed (http %d): %s", resp.StatusCode, msg)
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
		Code int
		Msg  string
		Data struct {
			State      string `json:"state"`
			ResultJSON string `json:"resultJson"`
			FailMsg    string `json:"failMsg"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", "", fmt.Errorf("stt poll: parse response: %w", err)
	}
	if parsed.Code != 0 && parsed.Code != 200 {
		return "", "", fmt.Errorf("stt poll failed (code %d): %s", parsed.Code, strings.TrimSpace(parsed.Msg))
	}
	if parsed.Data.State == "fail" {
		failMsg := strings.TrimSpace(parsed.Data.FailMsg)
		if failMsg == "" {
			failMsg = "transcription failed"
		}
		return "", "fail", fmt.Errorf("stt: %s", failMsg)
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
	if text := textFromMap(raw); text != "" {
		return text
	}
	if obj, ok := raw["resultObject"].(map[string]any); ok {
		if text := textFromMap(obj); text != "" {
			return text
		}
	}
	if transcripts, ok := raw["transcripts"].([]any); ok {
		var parts []string
		for _, item := range transcripts {
			if m, ok := item.(map[string]any); ok {
				if text := textFromMap(m); text != "" {
					parts = append(parts, text)
				}
			}
		}
		if len(parts) > 0 {
			return strings.TrimSpace(strings.Join(parts, "\n"))
		}
	}
	if urls, ok := raw["resultUrls"].([]any); ok && len(urls) > 0 {
		if u, ok := urls[0].(string); ok {
			return fetchTextURL(u)
		}
	}
	return ""
}

func textFromMap(raw map[string]any) string {
	for _, key := range []string{"text", "transcript", "transcription"} {
		if v, ok := raw[key].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
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
	content := strings.TrimSpace(string(body))
	if strings.HasPrefix(content, "{") {
		if text := extractTranscript(content); text != "" {
			return text
		}
	}
	return content
}
