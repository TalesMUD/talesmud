package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// BackgroundsHandler handles background image operations
type BackgroundsHandler struct {
	BasePath string // e.g., "./uploads/backgrounds"
}

// BackgroundInfo represents metadata about a background image
type BackgroundInfo struct {
	Name     string `json:"name"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

// Allowed image extensions
var allowedExtensions = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".webp": "image/webp",
}

// Maximum file size: 10MB
const maxFileSize = 10 * 1024 * 1024

// sanitizeFilename ensures the filename is safe and prevents path traversal
func sanitizeFilename(filename string) string {
	// Get base name (removes any path)
	filename = filepath.Base(filename)
	// Remove any remaining path separators
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	return filename
}

// isAllowedExtension checks if the file extension is allowed
func isAllowedExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	_, ok := allowedExtensions[ext]
	return ok
}

// getMimeType returns the MIME type for a given filename
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if mime, ok := allowedExtensions[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}

// ensureDir creates the base path directory if it doesn't exist
func (h *BackgroundsHandler) ensureDir() error {
	return os.MkdirAll(h.BasePath, 0755)
}

// ServeBackground serves a background image from disk
func (h *BackgroundsHandler) ServeBackground(c *gin.Context) {
	filename := c.Param("filename")
	filename = sanitizeFilename(filename)

	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename required"})
		return
	}

	filePath := filepath.Join(h.BasePath, filename)

	// Check if file exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "background not found"})
		return
	}
	if err != nil {
		log.WithError(err).Error("Error checking background file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Ensure it's a file, not a directory
	if info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	// Set content type and serve
	contentType := getMimeType(filename)
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=86400") // Cache for 1 day
	c.File(filePath)
}

// UploadBackground handles multipart upload of background images
func (h *BackgroundsHandler) UploadBackground(c *gin.Context) {
	// Ensure directory exists
	if err := h.ensureDir(); err != nil {
		log.WithError(err).Error("Failed to create backgrounds directory")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create storage directory"})
		return
	}

	// Limit request body size
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize)

	// Get the file from the form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.WithError(err).Error("Failed to get file from form")
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 10MB)"})
		return
	}

	// Get original filename and extension
	originalFilename := sanitizeFilename(header.Filename)
	ext := strings.ToLower(filepath.Ext(originalFilename))

	// Validate extension
	if !isAllowedExtension(originalFilename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type (allowed: png, jpg, jpeg, webp)"})
		return
	}

	// Get custom name from form or use original filename without extension
	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		name = strings.TrimSuffix(originalFilename, ext)
	}
	name = sanitizeFilename(name)

	// Build target filename (name + original extension)
	targetFilename := name + ext
	targetPath := filepath.Join(h.BasePath, targetFilename)

	// Create the target file
	out, err := os.Create(targetPath)
	if err != nil {
		log.WithError(err).Error("Failed to create target file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	defer out.Close()

	// Copy the uploaded file to the target
	if _, err := io.Copy(out, file); err != nil {
		log.WithError(err).Error("Failed to write file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	log.WithField("filename", targetFilename).Info("Background uploaded successfully")

	c.JSON(http.StatusOK, BackgroundInfo{
		Name:     name,
		Filename: targetFilename,
		URL:      fmt.Sprintf("/api/backgrounds/%s", targetFilename),
	})
}

// ListBackgrounds returns a list of all available background images
func (h *BackgroundsHandler) ListBackgrounds(c *gin.Context) {
	// Ensure directory exists
	if err := h.ensureDir(); err != nil {
		log.WithError(err).Error("Failed to create backgrounds directory")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to access storage directory"})
		return
	}

	entries, err := os.ReadDir(h.BasePath)
	if err != nil {
		log.WithError(err).Error("Failed to read backgrounds directory")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list backgrounds"})
		return
	}

	backgrounds := make([]BackgroundInfo, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !isAllowedExtension(filename) {
			continue
		}

		ext := filepath.Ext(filename)
		name := strings.TrimSuffix(filename, ext)

		backgrounds = append(backgrounds, BackgroundInfo{
			Name:     name,
			Filename: filename,
			URL:      fmt.Sprintf("/api/backgrounds/%s", filename),
		})
	}

	c.JSON(http.StatusOK, gin.H{"backgrounds": backgrounds})
}

// DeleteBackground removes a background image
func (h *BackgroundsHandler) DeleteBackground(c *gin.Context) {
	filename := c.Param("filename")
	filename = sanitizeFilename(filename)

	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename required"})
		return
	}

	if !isAllowedExtension(filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type"})
		return
	}

	filePath := filepath.Join(h.BasePath, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "background not found"})
		return
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		log.WithError(err).Error("Failed to delete background")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete background"})
		return
	}

	log.WithField("filename", filename).Info("Background deleted successfully")
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
