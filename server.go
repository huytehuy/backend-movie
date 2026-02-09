package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var movies = []Movie{
	{
		ID:          "1",
		Title:       "Sample Movie 1",
		Description: "This is a sample movie description",
		Thumbnail:   "/api/thumbnails/sample1.jpg",
		VideoURL:    "/api/videos/sample1.mp4",
		Duration:    7200,
		CreatedAt:   time.Now(),
	},
	{
		ID:          "2",
		Title:       "Sample Movie 2",
		Description: "Another sample movie",
		Thumbnail:   "/api/thumbnails/sample2.jpg",
		VideoURL:    "/api/videos/sample2.mp4",
		Duration:    5400,
		CreatedAt:   time.Now(),
	},
}

// GetMovies returns all movies
func GetMovies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

// GetMovie returns a single movie by ID
func GetMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	movieID := params["id"]

	for _, movie := range movies {
		if movie.ID == movieID {
			json.NewEncoder(w).Encode(movie)
			return
		}
	}

	http.Error(w, "Movie not found", http.StatusNotFound)
}

// StreamVideo handles video streaming with range support
func StreamVideo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	filename := params["filename"]

	// Security: prevent directory traversal
	if strings.Contains(filename, "..") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	videoPath := filepath.Join("videos", filename)

	// Check if file exists
	file, err := os.Open(videoPath)
	if err != nil {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Cannot get file info", http.StatusInternalServerError)
		return
	}

	fileSize := fileInfo.Size()

	// Handle Range requests for video seeking
	rangeHeader := r.Header.Get("Range")

	if rangeHeader == "" {
		// No range, send entire file
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
		w.Header().Set("Accept-Ranges", "bytes")
		io.Copy(w, file)
		return
	}

	// Parse range header
	ranges := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
	if len(ranges) != 2 {
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	start, err := strconv.ParseInt(ranges[0], 10, 64)
	if err != nil || start < 0 || start >= fileSize {
		http.Error(w, "Invalid range start", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	end := fileSize - 1
	if ranges[1] != "" {
		parsedEnd, err := strconv.ParseInt(ranges[1], 10, 64)
		if err == nil && parsedEnd < fileSize {
			end = parsedEnd
		}
	}

	contentLength := end - start + 1

	// Seek to start position
	_, err = file.Seek(start, 0)
	if err != nil {
		http.Error(w, "Seek error", http.StatusInternalServerError)
		return
	}

	// Set headers for partial content
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusPartialContent)

	// Copy the requested range
	io.CopyN(w, file, contentLength)
}

// ServeThumbnail serves video thumbnails
func ServeThumbnail(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	filename := params["filename"]

	if strings.Contains(filename, "..") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	thumbnailPath := filepath.Join("thumbnails", filename)

	// Check if file exists, if not serve a default
	if _, err := os.Stat(thumbnailPath); os.IsNotExist(err) {
		// Serve default thumbnail
		thumbnailPath = filepath.Join("thumbnails", "default.jpg")
	}

	http.ServeFile(w, r, thumbnailPath)
}

// UploadVideo handles video upload (optional)
func UploadVideo(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(200 << 20) // 200 MB max
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create videos directory if not exists
	os.MkdirAll("videos", os.ModePerm)

	// Create destination file
	dst, err := os.Create(filepath.Join("videos", handler.Filename))
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	log.Printf("Video uploaded: %s", handler.Filename)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Video uploaded successfully",
		"filename": handler.Filename,
	})
}

// HealthCheck endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
