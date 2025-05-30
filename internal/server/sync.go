// TODO gzip
package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"

	"zakirullin/stuffbot/internal/fs"
)

const (
	StorageDir            = "/app/mystorage"
	StatusOK              = "ok"
	StatusNotModified     = "not_modified"
	StatusUpdatedOnServer = "updated_on_server"
)

var (
	AuthToken string
)

var FS = func(userID int) *fs.FS {
	userFS, _ := fs.NewFS(StorageDir, afero.NewOsFs())
	return userFS
}

type file struct {
	Status       string `json:"status"`
	Path         string `json:"path"`
	LastModified int64  `json:"lastModified"`
	Content      string `json:"content"`
}

type syncRequest struct {
	Timestamps map[string]int64 `json:"timestamps"`
	Files      []file           `json:"files"` // New or modified files from client
}

type syncResponse struct {
	Status     string           `json:"status"`     // Status
	Files      []file           `json:"files"`      // Files with content that need syncing
	Timestamps map[string]int64 `json:"timestamps"` // Current server timestamps in Unix format
}

func SyncAllTextFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request syncRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Error parsing syncMediasRequest JSON: %v", err)
		http.Error(w, "Invalid syncMediasRequest JSON", http.StatusBadRequest)
		return
	}

	// 1) Save client-modified files to the server
	// 2) In case of conflict (server has a newer modification), merge the files and include them in the response
	// 3) Based on known client dirs timestamps, send newly updated or created files
	// 4) Respond with last modification timestamps for every dir

	// Save client-modified files to the server
	for _, clientFile := range request.Files {
		fullPath := filepath.Join(StorageDir, clientFile.Path)

		serverModTime := int64(0)
		// Check for any .../ attacks
		info, err := os.Stat(fullPath)
		if err == nil {
			serverModTime = info.ModTime().Unix()
		}
		var clientContent string

		if err != nil && !os.IsNotExist(err) {
			log.Printf("Error reading file '%s': %v", fullPath, err)
			logSync(fmt.Sprintf("Error reading file '%s': %v", fullPath, err))
			// All-or-nothing sync?
			continue
		} else if os.IsNotExist(err) {
			logSync(fmt.Sprintf("Creating: '%s'", clientFile.Path))
			clientContent = clientFile.Content
		} else {
			// file locks?
			fileWasModifiedOnServer := serverModTime > clientFile.LastModified
			if fileWasModifiedOnServer {
				serverContent, err := ioutil.ReadFile(fullPath)
				if err != nil {
					log.Printf("Error reading file '%s': %v", fullPath, err)
					continue
				}
				logSync(fmt.Sprintf("Merging and writing: '%s'", clientFile.Path))
				clientContent = Merge(string(serverContent), clientFile.Content)
			} else {
				// Server file hasn't changed since client's last sync
				logSync(fmt.Sprintf("Writing only: '%s'", clientFile.Path))
				clientContent = clientFile.Content
			}
		}

		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			log.Printf("Error creating directory for file '%s': %v", fullPath, err)
			logSync(fmt.Sprintf("Error creating directory for file '%s': %v", fullPath, err))
			continue
		}

		// Write the clientContent to the server at path
		err = os.WriteFile(fullPath, []byte(clientContent), 0644)
		if err != nil {
			log.Printf("Error writing file '%s': %v", fullPath, err)
			logSync(fmt.Sprintf("Error writing file '%s': %v", fullPath, err))
			continue
		}
	}

	serverTimestamps, err := timestamps(StorageDir)
	if err != nil {
		log.Printf("Error getting server timestamps: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get timestamps: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare the list of files to send to the client
	// TODO optimize don't send files known to client.
	// For now we save client file to server, and the code below would include it again.
	files := make([]file, 0)
	dirTimestamps := make(map[string]int64)
	for path, serverFileTime := range serverTimestamps {
		parts := strings.Split(path, string(os.PathSeparator))
		dir := parts[0]
		isInRoot := len(parts) == 1
		if isInRoot {
			dir = "."
		}

		requestDirTime, exists := request.Timestamps[dir]
		if !exists || serverFileTime > requestDirTime {
			// Client needs this file - read its content
			fullPath := filepath.Join(StorageDir, path)
			content, err := ioutil.ReadFile(fullPath)
			if err != nil {
				log.Printf("Error reading file %s: %v", fullPath, err)
				continue
			}
			logSync(fmt.Sprintf("Sending file: '%s'", path))

			files = append(files, file{
				Status:       StatusOK,
				Path:         path,
				LastModified: serverFileTime,
				Content:      string(content),
			})
		}

		// Calculate the latest file timestamp for each directory
		existingTimestamp, exists := dirTimestamps[dir]
		if !exists {
			dirTimestamps[dir] = serverFileTime
			continue
		}
		if serverFileTime > existingTimestamp {
			dirTimestamps[dir] = serverFileTime
		}
	}

	response := syncResponse{
		Status:     StatusOK,
		Files:      files,
		Timestamps: dirTimestamps,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding sync response: %v", err)
	}
}

func SyncText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var clientFile file
	if err := json.NewDecoder(r.Body).Decode(&clientFile); err != nil {
		log.Printf("Error parsing syncMediasRequest JSON: %v", err)
		http.Error(w, "Invalid syncMediasRequest JSON", http.StatusBadRequest)
		return
	}

	// 1) Save client-modified file to the server
	// 2) In case of conflict (server has a newer modification), merge the clientFile and include them in the response

	// Save client-modified clientFile to the server
	fullPath := filepath.Join(StorageDir, clientFile.Path)

	serverModTime := int64(0)
	// Check for any .../ attacks
	info, err := os.Stat(fullPath)
	if err == nil {
		serverModTime = info.ModTime().Unix()
	}

	// TODO if no clientFile, severContent = ""
	serverContent, err := ioutil.ReadFile(fullPath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Error reading one clientFile '%s': %v", fullPath, err)
		http.Error(w, "Error reading server clientFile", http.StatusBadRequest)
		return
	}

	// TODO when clientFile does not exist the content is empty, which is implicit
	// Return already up-to-date status
	if string(serverContent) == clientFile.Content {
		response := map[string]interface{}{
			"status":       StatusNotModified,
			"lastModified": serverModTime,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotModified)
		json.NewEncoder(w).Encode(response)
		return
	}

	var content string
	fileWasModifiedOnServer := false
	if os.IsNotExist(err) {
		logSync(fmt.Sprintf("Creating one clientFile: '%s'", clientFile.Path))
		content = clientFile.Content
	} else {
		fileWasModifiedOnServer = serverModTime > clientFile.LastModified
		if fileWasModifiedOnServer {
			logSync(fmt.Sprintf("Server one clientFile '%s' was modified at %d, client timestamp is %d", fullPath, serverModTime, clientFile.LastModified))
			logSync(fmt.Sprintf("Merging and writing one clientFile: '%s'", clientFile.Path))
			content = Merge(string(serverContent), clientFile.Content)
		} else {
			// TODO for resilience add merge here, because we had case when server saved latest TS but no conent.
			// Also, if for some reason timestamps would change on server migration and such.
			// Server clientFile hasn't changed since client's last sync
			logSync(fmt.Sprintf("Writing only one clientFile: '%s'", clientFile.Path))
			content = clientFile.Content
		}
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		log.Printf("Error creating directory for clientFile '%s': %v", fullPath, err)
		logSync(fmt.Sprintf("Error creating directory for clientFile '%s': %v", fullPath, err))
		http.Error(w, "Error creating directory", http.StatusInternalServerError)
		return
	}

	// Write the content to the server at path
	err = os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		log.Printf("Error writing clientFile '%s': %v", fullPath, err)
		logSync(fmt.Sprintf("Error writing clientFile '%s': %v", fullPath, err))
		http.Error(w, "Error writing clientFile", http.StatusInternalServerError)
		return
	}
	info, err = os.Stat(fullPath)
	if err == nil {
		serverModTime = info.ModTime().Unix()
	}
	logSync(fmt.Sprintf("Server timestamp for '%s': %d", fullPath, serverModTime))

	if !fileWasModifiedOnServer {
		response := map[string]interface{}{
			"status":       StatusUpdatedOnServer,
			"lastModified": serverModTime,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get clientFile timestamp
	info, err = os.Stat(fullPath)
	if err != nil {
		log.Printf("Error getting one clientFile '%s' timestamp: %v", fullPath, err)
		logSync(fmt.Sprintf("Error getting one clientFile '%s' timestamp: %v", fullPath, err))
		http.Error(w, "Error getting clientFile timestamp", http.StatusInternalServerError)
		return
	}

	response := file{
		Status:       StatusOK,
		Content:      content,
		Path:         clientFile.Path,
		LastModified: serverModTime,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding sync response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// validateAuthToken checks if the syncMediasRequest has a valid auth token
func validateAuthToken(r *http.Request) bool {
	token := r.Header.Get("Authorization")

	if strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}

	return token == AuthToken
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateAuthToken(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// timestamps recursively scans a directory and returns the latest modification time
// for directories only (including root directory) as Unix timestamps
func timestamps(rootPath string) (map[string]int64, error) {
	timestamps := make(map[string]int64)
	realPath, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		log.Printf("Warning: Could not resolve symlink: %v. Using original path.", err)
		realPath = rootPath
	}

	err = filepath.Walk(realPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") && path != realPath {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, err := filepath.Rel(realPath, path)
		if err != nil {
			return nil
		}

		// Skip non-markdown files for file processing
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		if relPath == "" {
			relPath = "."
		}

		timestamps[relPath] = info.ModTime().Unix()

		return nil
	})

	if err != nil {
		return nil, err
	}

	return timestamps, nil
}

func logSync(msg string) {
	file, err := os.OpenFile("/tmp/sync", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer file.Close()

	time := time.Now().Format("2006-01-02 15:04:05")
	if _, err := file.WriteString(time + ": " + msg + "\n"); err != nil {
		fmt.Println("Error writing to log file:", err)
		return
	}
}
