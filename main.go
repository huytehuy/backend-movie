package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Create necessary directories
	os.MkdirAll("videos", os.ModePerm)
	os.MkdirAll("thumbnails", os.ModePerm)

	// Initialize router
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Movie routes
	api.HandleFunc("/movies", GetMovies).Methods("GET")
	api.HandleFunc("/movies/{id}", GetMovie).Methods("GET")
	api.HandleFunc("/upload", UploadVideo).Methods("POST")

	// Video streaming routes
	api.HandleFunc("/videos/{filename}", StreamVideo).Methods("GET")
	api.HandleFunc("/thumbnails/{filename}", ServeThumbnail).Methods("GET")

	// Watch party routes
	api.HandleFunc("/rooms", CreateRoom).Methods("POST")
	api.HandleFunc("/rooms", GetActiveRooms).Methods("GET")
	api.HandleFunc("/rooms/{id}", GetRoom).Methods("GET")
	api.HandleFunc("/rooms/{id}/ws", HandleWebSocket)

	// Health check
	api.HandleFunc("/health", HealthCheck).Methods("GET")

	// Serve React build (production)
	// Uncomment this when you build your React app
	// router.PathPrefix("/").Handler(http.FileServer(http.Dir("./movie/dist")))

	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start room cleanup routine
	StartRoomCleanup()

	log.Printf("Server starting on port %s", port)
	log.Printf("Video streaming: http://localhost:%s/api/videos/", port)
	log.Printf("WebSocket: ws://localhost:%s/api/rooms/{roomId}/ws", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
