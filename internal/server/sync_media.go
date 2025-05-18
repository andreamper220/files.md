package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"zakirullin/stuffbot/internal/fs"
)

var syncMediasRequest struct {
	Dir           string `json:"dir"`
	Timestamp     int64  `json:"timestamp"`
	FilenamesHash string `json:"filenamesHash"`
}

type media struct {
	Path         string `json:"path"`
	LastModified int64  `json:"lastModified"`
}

func SyncAllMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&syncMediasRequest); err != nil {
		log.Printf("Error parsing syncMediasRequest JSON: %v", err)
		http.Error(w, "Invalid syncMediasRequest JSON", http.StatusBadRequest)
		return
	}

	// TODO ../.. Attacks
	mediaFolder := filepath.Join(StorageDir, fs.DirImg)
	logSync(fmt.Sprintf("Media sync syncMediasRequest for folder: '%s', last sync: %d", syncMediasRequest.Dir, syncMediasRequest.Timestamp))

	if _, err := os.Stat(mediaFolder); os.IsNotExist(err) {
		emptyResponse := struct {
			Files     []interface{} `json:"files"`
			Timestamp int64         `json:"timestamp"`
		}{
			Files:     []interface{}{},
			Timestamp: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(emptyResponse)
		return
	}

	var mediaFiles []media
	latestTimestamp := int64(0)

	// Find media files newer than client's timestamp
	err := filepath.Walk(mediaFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		modTime := info.ModTime().Unix()
		if modTime <= syncMediasRequest.Timestamp {
			return nil
		}
		if modTime > latestTimestamp {
			latestTimestamp = modTime
		}

		relPath, err := filepath.Rel(mediaFolder, path)
		if err != nil {
			return nil
		}

		mediaFiles = append(mediaFiles, media{
			Path:         relPath,
			LastModified: modTime,
		})

		return nil
	})

	if err != nil {
		log.Printf("Error scanning media folder: %v", err)
		http.Error(w, "Error scanning media folder", http.StatusInternalServerError)
		return
	}

	response := struct {
		Files     []media `json:"files"`
		Timestamp int64   `json:"timestamp"`
	}{
		Files:     mediaFiles,
		Timestamp: latestTimestamp,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding media sync response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func SyncMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// decode media from json
	var serverMedia media
	if err := json.NewDecoder(r.Body).Decode(&serverMedia); err != nil {
		log.Printf("Error parsing syncMedia Request JSON: %v", err)
		http.Error(w, "Invalid syncMedia Request JSON", http.StatusBadRequest)
		return
	}

	// TODO ../.. attacks
	filePath := filepath.Join(StorageDir, fs.DirImg, serverMedia.Path)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
}
