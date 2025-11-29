package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"uploads/internal/config"
	"uploads/internal/database"
)

type Handler struct {
	config *config.Config
	db     *database.DB
}

func NewHandler(cfg *config.Config, db *database.DB) *Handler {
	return &Handler{config: cfg, db: db}
}

// Login handler
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Create MD5 hash of submitted password
	hasher := md5.New()
	hasher.Write([]byte(creds.Password))
	submittedHash := hex.EncodeToString(hasher.Sum(nil))

	if creds.Username == h.config.UIUsername && submittedHash == h.config.UIPasswordHash {
		// Login successful, return API key for UI use
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"api_key": h.config.APIKey,
		})
		return
	}

	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}

// UploadFile handler
func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form - limit to 10MB
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate new filename with timestamp only
	extension := filepath.Ext(handler.Filename)
	timestamp := time.Now().Unix()
	newFilename := fmt.Sprintf("%d%s", timestamp, extension)
	
	// Get file size by reading it into memory temporarily
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Save file with new name
	filePath := filepath.Join(h.config.DataDir, newFilename)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file on server: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := dst.Write(fileBytes); err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert file record into database
	fileSize := int64(len(fileBytes))
	uploadAddr := r.RemoteAddr
	if err := h.db.InsertFile(handler.Filename, newFilename, fileSize, uploadAddr); err != nil {
		// If database insertion fails, we should clean up the file
		os.Remove(filePath)
		http.Error(w, "Failed to save file record to database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	
	// Create response with filename and public URL
	response := map[string]interface{}{
		"message": fmt.Sprintf("File '%s' successfully uploaded.", newFilename),
		"filename": newFilename,
		"url": fmt.Sprintf("https://%s/file/%s", r.Host, newFilename),
	}
	
	json.NewEncoder(w).Encode(response)
}

// ListFiles handler
func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all file records from database
	fileRecords, err := h.db.GetAllFiles()
	if err != nil {
		http.Error(w, "Failed to read file records from database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// For debugging, let's make sure we have valid data
	if fileRecords == nil {
	fileRecords = []*database.FileRecord{} // Return empty array instead of null
	}

	// Convert to a list of file records for response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileRecords)
}

// DownloadFile handler
func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("name")
	if filename == "" {
		http.Error(w, "Missing 'name' parameter", http.StatusBadRequest)
		return
	}

	// Prevent Path Traversal Attack
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.config.DataDir, filename)

	// Ensure file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// For HEAD requests, only return headers without the file content
	if r.Method == http.MethodHead {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			http.Error(w, "Failed to get file info", http.StatusInternalServerError)
			return
		}
		
		// Set headers for HEAD request
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		return
	}

	// For GET requests, serve the file
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	http.ServeFile(w, r, filePath)
}

// PublicFileAccess handler for public file access at /file/{filename}
func (h *Handler) PublicFileAccess(w http.ResponseWriter, r *http.Request) {
	// Extract filename from URL path
	filename := strings.TrimPrefix(r.URL.Path, "/file/")
	
	if filename == "" {
		http.Error(w, "Filename not specified", http.StatusBadRequest)
		return
	}

	// Prevent Path Traversal Attack
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
	http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.config.DataDir, filename)

	// Ensure file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
	http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// For HEAD requests, only return headers without the file content
	if r.Method == http.MethodHead {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			http.Error(w, "Failed to get file info", http.StatusInternalServerError)
			return
		}
		
		// Set headers for HEAD request
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
		w.Header().Set("Content-Disposition", "inline; filename="+filename)
		return
	}

	// For GET requests, serve the file
	w.Header().Set("Content-Disposition", "inline; filename="+filename)
	http.ServeFile(w, r, filePath)
}

// DeleteFile handler
func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Filename string `json:"filename"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Prevent Path Traversal Attack
	if strings.Contains(req.Filename, "..") || strings.Contains(req.Filename, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.config.DataDir, req.Filename)

	// Check if file exists in filesystem
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Delete file from filesystem
	if err := os.Remove(filePath); err != nil {
		http.Error(w, "Failed to delete file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete file record from database
	if err := h.db.DeleteFile(req.Filename); err != nil {
		// Log the error but don't fail the request since the file was already deleted from filesystem
		// This might happen if the file wasn't properly recorded in the database
		fmt.Printf("Warning: Failed to delete file record from database: %v\n", err)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File '%s' successfully deleted.", req.Filename)
}
