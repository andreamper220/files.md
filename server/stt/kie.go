// Package stt transcribes audio via kie.ai (ElevenLabs speech-to-text).
// API docs: https://docs.kie.ai/market/quickstart
package stt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"slices"
	"strings"
	"time"
)

const (
	// File uploads: https://docs.kie.ai/file-upload-api/quickstart
	streamUploadURL = "https://kieai.redpandaai.co/api/file-stream-upload"
	base64UploadURL = "https://kieai.redpandaai.co/api/file-base64-upload"
	urlUploadURL    = "https://kieai.redpandaai.co/api/file-url-upload"
	uploadPath      = "filesmd/voice"

	// Market jobs: https://docs.kie.ai/market/quickstart
	createTaskURL = "https://api.kie.ai/api/v1/jobs/createTask"
	recordInfoURL = "https://api.kie.ai/api/v1/jobs/recordInfo"
	model         = "elevenlabs/speech-to-text"

	maxBase64UploadBytes = 10 * 1024 * 1024
)

// transcribeKie uploads audio to kie.ai and returns the transcript text.
func transcribeKie(apiKey string, audio []byte, mimeType, languageCode string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("stt: KIE_API_KEY is not set")
	}
	if len(audio) == 0 {
		return "", fmt.Errorf("stt: empty audio")
	}
	if mimeType == "" {
		mimeType = "audio/ogg"
	}

	audioURLs, err := uploadAudio(apiKey, audio, mimeType)
	if err != nil {
		return "", err
	}

	// Give the upload CDN a moment before kie.ai fetches the file.
	time.Sleep(800 * time.Millisecond)

	taskID, err := createTask(apiKey, audioURLs, languageCode)
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

func uploadAudio(apiKey string, audio []byte, mimeType string) ([]string, error) {
	var collected []string
	var lastErr error

	// Base64 upload returns stable kieai.redpandaai.co URLs more often than stream/tempfile.
	if len(audio) <= maxBase64UploadBytes {
		if urls, err := uploadAudioBase64(apiKey, audio, mimeType); err == nil {
			collected = append(collected, urls...)
		} else {
			lastErr = err
		}
	}
	if urls, err := uploadAudioStream(apiKey, audio, mimeType); err == nil {
		collected = append(collected, urls...)
	} else if lastErr == nil {
		lastErr = err
	}
	if len(collected) == 0 {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, fmt.Errorf("stt upload: no download URL in response")
	}

	sttURLs, err := prepareSTTAudioURLs(apiKey, uniqueURLs(collected...), mimeType)
	if err != nil {
		return nil, err
	}
	if len(sttURLs) == 0 {
		return nil, fmt.Errorf("stt upload: no STT-ready audio URL")
	}
	return sttURLs, nil
}

func prepareSTTAudioURLs(apiKey string, urls []string, mimeType string) ([]string, error) {
	ordered := orderURLsForSTT(urls)
	seen := map[string]bool{}
	var out []string
	add := func(u string) {
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			return
		}
		seen[u] = true
		out = append(out, u)
	}

	for _, u := range ordered {
		if isTempfileURL(u) {
			rehosted, err := uploadAudioFromURL(apiKey, u, mimeType)
			if err != nil {
				slog.Warn("stt rehost tempfile url failed", "url", u, "err", err)
				continue
			}
			for _, candidate := range orderURLsForSTT(rehosted) {
				if isTempfileURL(candidate) {
					continue
				}
				add(candidate)
			}
			continue
		}
		add(u)
	}
	if len(out) == 0 {
		for _, u := range ordered {
			add(u)
		}
	}
	return out, nil
}

func isTempfileURL(u string) bool {
	return strings.Contains(strings.ToLower(u), "tempfile.redpandaai.co")
}

func orderURLsForSTT(urls []string) []string {
	score := func(u string) int {
		lower := strings.ToLower(u)
		switch {
		case strings.Contains(lower, "kieai.redpandaai.co/download"):
			return 0
		case strings.Contains(lower, "kieai.redpandaai.co/files"):
			return 1
		case strings.Contains(lower, "kieai.redpandaai.co"):
			return 2
		case strings.Contains(lower, "tempfile.redpandaai.co"):
			return 4
		default:
			return 3
		}
	}
	out := append([]string(nil), urls...)
	slices.SortStableFunc(out, func(a, b string) int {
		return score(a) - score(b)
	})
	return out
}

func uploadAudioStream(apiKey string, audio []byte, mimeType string) ([]string, error) {
	raw, err := uploadAudioStreamRaw(apiKey, audio, mimeType)
	if err != nil {
		return nil, err
	}
	return parseUploadJSON(raw)
}

func uploadAudioBase64(apiKey string, audio []byte, mimeType string) ([]string, error) {
	raw, err := uploadAudioBase64Raw(apiKey, audio, mimeType)
	if err != nil {
		return nil, err
	}
	return parseUploadJSON(raw)
}

func uploadAudioFromURL(apiKey, sourceURL, mimeType string) ([]string, error) {
	raw, err := uploadAudioFromURLRaw(apiKey, sourceURL, mimeType)
	if err != nil {
		return nil, err
	}
	return parseUploadJSON(raw)
}

// DebugUploadResponses returns raw upload API responses for troubleshooting.
func DebugUploadResponses(apiKey string, audio []byte, mimeType string) (streamJSON, base64JSON, rehostJSON string) {
	if raw, err := uploadAudioStreamRaw(apiKey, audio, mimeType); err == nil {
		streamJSON = raw
	} else {
		streamJSON = err.Error()
	}
	if len(audio) <= maxBase64UploadBytes {
		if raw, err := uploadAudioBase64Raw(apiKey, audio, mimeType); err == nil {
			base64JSON = raw
		} else {
			base64JSON = err.Error()
		}
	}
	if streamJSON != "" {
		var parsed struct {
			Data struct {
				FileURL     string `json:"fileUrl"`
				DownloadURL string `json:"downloadUrl"`
			} `json:"data"`
		}
		if json.Unmarshal([]byte(streamJSON), &parsed) == nil {
			source := strings.TrimSpace(parsed.Data.DownloadURL)
			if source == "" {
				source = strings.TrimSpace(parsed.Data.FileURL)
			}
			if source != "" {
				if raw, err := uploadAudioFromURLRaw(apiKey, source, mimeType); err == nil {
					rehostJSON = raw
				} else {
					rehostJSON = err.Error()
				}
			}
		}
	}
	return streamJSON, base64JSON, rehostJSON
}

func uploadAudioStreamRaw(apiKey string, audio []byte, mimeType string) (string, error) {
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

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("stt upload: read response: %w", err)
	}
	return string(raw), nil
}

func uploadAudioBase64Raw(apiKey string, audio []byte, mimeType string) (string, error) {
	fileName := fmt.Sprintf("voice_%d%s", time.Now().UnixNano(), audioExt(mimeType))
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(audio))
	body, _ := json.Marshal(map[string]string{
		"base64Data": dataURL,
		"uploadPath": uploadPath,
		"fileName":   fileName,
	})

	req, err := http.NewRequest(http.MethodPost, base64UploadURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("stt base64 upload: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("stt base64 upload: read response: %w", err)
	}
	return string(raw), nil
}

func uploadAudioFromURLRaw(apiKey, sourceURL, mimeType string) (string, error) {
	fileName := fmt.Sprintf("voice_%d%s", time.Now().UnixNano(), audioExt(mimeType))
	body, _ := json.Marshal(map[string]string{
		"fileUrl":    sourceURL,
		"uploadPath": uploadPath,
		"fileName":   fileName,
	})

	req, err := http.NewRequest(http.MethodPost, urlUploadURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("stt url upload: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("stt url upload: read response: %w", err)
	}
	return string(raw), nil
}

func parseUploadResponse(respBody io.Reader) ([]string, error) {
	raw, err := io.ReadAll(respBody)
	if err != nil {
		return nil, err
	}
	return parseUploadJSON(string(raw))
}

func parseUploadJSON(raw string) ([]string, error) {
	var parsed struct {
		Success bool `json:"success"`
		Code    int  `json:"code"`
		Msg     string
		Data    struct {
			DownloadURL string `json:"downloadUrl"`
			FileURL     string `json:"fileUrl"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("stt upload: parse response: %w", err)
	}
	// downloadUrl is often the link kie.ai can fetch; fileUrl may redirect or require auth.
	urls := uniqueURLs(parsed.Data.DownloadURL, parsed.Data.FileURL)
	if parsed.Code != 200 || !parsed.Success || len(urls) == 0 {
		msg := strings.TrimSpace(parsed.Msg)
		if msg == "" {
			msg = "no download URL in response"
		}
		if parsed.Code != 0 && parsed.Code != 200 {
			return nil, fmt.Errorf("stt upload failed (code %d): %s", parsed.Code, msg)
		}
		return nil, fmt.Errorf("stt upload failed: %s", msg)
	}
	return urls, nil
}

func uniqueURLs(urls ...string) []string {
	seen := map[string]bool{}
	var out []string
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, u)
	}
	return out
}

func createTask(apiKey string, audioURLs []string, languageCode string) (string, error) {
	languageCode = strings.TrimSpace(languageCode)
	if languageCode == "" {
		languageCode = "ru"
	}

	var lastErr error
	for _, audioURL := range audioURLs {
		if !isAudioURLReachable(audioURL) {
			slog.Warn("stt audio url not reachable from bot host, trying anyway", "audio_url", audioURL)
		}
		for _, input := range sttInputVariants(audioURL, languageCode) {
			body, _ := json.Marshal(map[string]any{
				"model": model,
				"input": input,
			})
			for attempt := 1; attempt <= 3; attempt++ {
				taskID, err := createTaskOnce(apiKey, body)
				if err == nil {
					return taskID, nil
				}
				lastErr = err
				slog.Warn("stt create task attempt failed",
					"attempt", attempt,
					"audio_url", audioURL,
					"language_code", input["language_code"],
					"err", err,
				)
				if attempt < 3 && isRetryableSTTError(err) {
					time.Sleep(time.Duration(attempt) * time.Second)
					continue
				}
				break
			}
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("stt create task: no audio URLs to try")
	}
	return "", lastErr
}

// sttInputVariants returns createTask input payloads to try, ordered by API docs preference.
func sttInputVariants(audioURL, languageCode string) []map[string]any {
	return []map[string]any{
		{"audio_url": audioURL},
		{
			"audio_url":        audioURL,
			"language_code":    languageCode,
			"tag_audio_events": false,
			"diarize":          false,
		},
		{
			"audio_url":        audioURL,
			"language_code":    languageCode,
			"tag_audio_events": true,
			"diarize":          false,
		},
		{
			"audio_url":        audioURL,
			"language_code":    "",
			"tag_audio_events": false,
			"diarize":          false,
		},
	}
}

func isAudioURLReachable(audioURL string) bool {
	client := &http.Client{Timeout: 10 * time.Second}
	for _, method := range []string{http.MethodHead, http.MethodGet} {
		req, err := http.NewRequest(method, audioURL, nil)
		if err != nil {
			return false
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			return true
		}
	}
	return false
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
		msg = strings.TrimSpace(string(respBody))
	}
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
