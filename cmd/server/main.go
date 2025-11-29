package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"uploads/internal/config"
	"uploads/internal/database"
	"uploads/internal/handlers"
	"uploads/internal/middleware"
)

func main() {
	// 1. Load configuration
	cfg, err := config.LoadConfig("configs/config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize database
	db, err := database.NewDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize handlers with config and database
	handler := handlers.NewHandler(cfg, db)

	// 2. Setup routes with middleware
	http.HandleFunc("/api/upload", middleware.CORSMiddleware(middleware.APIKeyMiddleware(cfg.APIKey, handler.UploadFile)))
	http.HandleFunc("/api/list", middleware.CORSMiddleware(middleware.APIKeyMiddleware(cfg.APIKey, handler.ListFiles)))
	http.HandleFunc("/api/download", middleware.CORSMiddleware(middleware.APIKeyMiddleware(cfg.APIKey, handler.DownloadFile)))
	http.HandleFunc("/api/delete", middleware.CORSMiddleware(middleware.APIKeyMiddleware(cfg.APIKey, handler.DeleteFile)))
	http.HandleFunc("/api/login", middleware.CORSMiddleware(handler.Login)) // No API key required for login

	// 2.5 Public file access endpoint
	http.HandleFunc("/file/", handler.PublicFileAccess) // No API key required for public access

	// 3. Serve static files
	http.Handle("/", http.FileServer(http.Dir("assets/")))

	// 4. Ensure data directory exists
	if err := os.MkdirAll(filepath.Join(".", cfg.DataDir), 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// 5. Start server
	log.Printf("Server running on http://localhost:%d", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.ServerAddr(), nil))
}
