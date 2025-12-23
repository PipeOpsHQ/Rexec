package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rexec/rexec/internal/models"
	"github.com/rexec/rexec/internal/storage"
)

// TutorialHandler handles tutorial API endpoints
type TutorialHandler struct {
	store   *storage.PostgresStore
	s3Store *storage.S3Store
}

// NewTutorialHandler creates a new TutorialHandler
func NewTutorialHandler(store *storage.PostgresStore) *TutorialHandler {
	handler := &TutorialHandler{
		store: store,
	}

	// Initialize S3 store if configured
	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket != "" {
		s3Config := storage.S3Config{
			Bucket:          s3Bucket,
			Region:          os.Getenv("S3_REGION"),
			Endpoint:        os.Getenv("S3_ENDPOINT"),
			AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
			Prefix:          os.Getenv("S3_PREFIX"),
			ForcePathStyle:  os.Getenv("S3_FORCE_PATH_STYLE") == "true",
		}

		s3Store, err := storage.NewS3Store(s3Config)
		if err != nil {
			log.Printf("[Tutorial] Warning: Failed to initialize S3 store: %v (image uploads will be disabled)", err)
		} else {
			handler.s3Store = s3Store
			log.Printf("[Tutorial] Handler initialized (storing images in S3: %s)", s3Bucket)
		}
	}

	return handler
}

// CreateTutorialRequest represents the request to create a tutorial
type CreateTutorialRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	VideoURL    string `json:"video_url"`
	Thumbnail   string `json:"thumbnail"`
	Duration    string `json:"duration"`
	Category    string `json:"category"`
	Order       int    `json:"order"`
	IsPublished bool   `json:"is_published"`
}

// UpdateTutorialRequest represents the request to update a tutorial
type UpdateTutorialRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Type        *string `json:"type"`
	Content     *string `json:"content"`
	VideoURL    *string `json:"video_url"`
	Thumbnail   *string `json:"thumbnail"`
	Duration    *string `json:"duration"`
	Category    *string `json:"category"`
	Order       *int    `json:"order"`
	IsPublished *bool   `json:"is_published"`
}

// TutorialResponse represents a tutorial in API responses
type TutorialResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Content     string    `json:"content"`
	VideoURL    string    `json:"video_url"`
	Thumbnail   string    `json:"thumbnail"`
	Duration    string    `json:"duration"`
	Category    string    `json:"category"`
	Order       int       `json:"order"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toTutorialResponse(t *models.Tutorial) TutorialResponse {
	return TutorialResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Type:        t.Type,
		Content:     t.Content,
		VideoURL:    t.VideoURL,
		Thumbnail:   t.Thumbnail,
		Duration:    t.Duration,
		Category:    t.Category,
		Order:       t.Order,
		IsPublished: t.IsPublished,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// ListPublicTutorials lists all published tutorials (public endpoint)
// GET /api/tutorials
func (h *TutorialHandler) ListPublicTutorials(c *gin.Context) {
	category := c.Query("category")

	tutorials, err := h.store.GetPublishedTutorials(c.Request.Context(), category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{Code: http.StatusInternalServerError, Message: "failed to fetch tutorials"})
		return
	}

	response := make([]TutorialResponse, 0, len(tutorials))
	for _, t := range tutorials {
		response = append(response, toTutorialResponse(t))
	}

	// Group by category for frontend convenience
	categories := make(map[string][]TutorialResponse)
	for _, t := range response {
		categories[t.Category] = append(categories[t.Category], t)
	}

	c.JSON(http.StatusOK, gin.H{
		"tutorials":  response,
		"categories": categories,
		"count":      len(response),
	})
}

// ListAllTutorials lists all tutorials including unpublished (admin only)
// GET /api/admin/tutorials
func (h *TutorialHandler) ListAllTutorials(c *gin.Context) {
	tutorials, err := h.store.GetAllTutorials(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{Code: http.StatusInternalServerError, Message: "failed to fetch tutorials"})
		return
	}

	response := make([]TutorialResponse, 0, len(tutorials))
	for _, t := range tutorials {
		response = append(response, toTutorialResponse(t))
	}

	c.JSON(http.StatusOK, gin.H{
		"tutorials": response,
		"count":     len(response),
	})
}

// GetTutorial retrieves a single tutorial by ID
// GET /api/tutorials/:id
func (h *TutorialHandler) GetTutorial(c *gin.Context) {
	id := c.Param("id")

	tutorial, err := h.store.GetTutorialByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIError{Code: http.StatusNotFound, Message: "tutorial not found"})
		return
	}

	// Only return published tutorials for public endpoint
	// Admin endpoint uses ListAllTutorials
	if !tutorial.IsPublished {
		c.JSON(http.StatusNotFound, models.APIError{Code: http.StatusNotFound, Message: "tutorial not found"})
		return
	}

	c.JSON(http.StatusOK, toTutorialResponse(tutorial))
}

// CreateTutorial creates a new tutorial (admin only)
// POST /api/admin/tutorials
func (h *TutorialHandler) CreateTutorial(c *gin.Context) {
	var req CreateTutorialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIError{Code: http.StatusBadRequest, Message: "invalid request: " + err.Error()})
		return
	}

	now := time.Now()
	tutorial := &models.Tutorial{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Content:     req.Content,
		VideoURL:    req.VideoURL,
		Thumbnail:   req.Thumbnail,
		Duration:    req.Duration,
		Category:    req.Category,
		Order:       req.Order,
		IsPublished: req.IsPublished,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if tutorial.Type == "" {
		tutorial.Type = "video"
	}

	if tutorial.Category == "" {
		tutorial.Category = "getting-started"
	}

	if err := h.store.CreateTutorial(c.Request.Context(), tutorial); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{Code: http.StatusInternalServerError, Message: "failed to create tutorial"})
		return
	}

	c.JSON(http.StatusCreated, toTutorialResponse(tutorial))
}

// UpdateTutorial updates an existing tutorial (admin only)
// PUT /api/admin/tutorials/:id
func (h *TutorialHandler) UpdateTutorial(c *gin.Context) {
	id := c.Param("id")

	tutorial, err := h.store.GetTutorialByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIError{Code: http.StatusNotFound, Message: "tutorial not found"})
		return
	}

	var req UpdateTutorialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIError{Code: http.StatusBadRequest, Message: "invalid request: " + err.Error()})
		return
	}

	// Update fields if provided
	if req.Title != nil {
		tutorial.Title = *req.Title
	}
	if req.Description != nil {
		tutorial.Description = *req.Description
	}
	if req.Type != nil {
		tutorial.Type = *req.Type
	}
	if req.Content != nil {
		tutorial.Content = *req.Content
	}
	if req.VideoURL != nil {
		tutorial.VideoURL = *req.VideoURL
	}
	if req.Thumbnail != nil {
		tutorial.Thumbnail = *req.Thumbnail
	}
	if req.Duration != nil {
		tutorial.Duration = *req.Duration
	}
	if req.Category != nil {
		tutorial.Category = *req.Category
	}
	if req.Order != nil {
		tutorial.Order = *req.Order
	}
	if req.IsPublished != nil {
		tutorial.IsPublished = *req.IsPublished
	}
	tutorial.UpdatedAt = time.Now()

	if err := h.store.UpdateTutorial(c.Request.Context(), tutorial); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{Code: http.StatusInternalServerError, Message: "failed to update tutorial"})
		return
	}

	c.JSON(http.StatusOK, toTutorialResponse(tutorial))
}

// DeleteTutorial deletes a tutorial (admin only)
// DELETE /api/admin/tutorials/:id
func (h *TutorialHandler) DeleteTutorial(c *gin.Context) {
	id := c.Param("id")

	// Verify tutorial exists
	_, err := h.store.GetTutorialByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIError{Code: http.StatusNotFound, Message: "tutorial not found"})
		return
	}

	if err := h.store.DeleteTutorial(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{Code: http.StatusInternalServerError, Message: "failed to delete tutorial"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tutorial deleted"})
}

// UploadImage uploads an image for a tutorial/guide
// POST /api/admin/tutorials/images
func (h *TutorialHandler) UploadImage(c *gin.Context) {
	if h.s3Store == nil {
		c.JSON(http.StatusServiceUnavailable, models.APIError{Code: http.StatusServiceUnavailable, Message: "storage not configured"})
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIError{Code: http.StatusBadRequest, Message: "invalid file"})
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		c.JSON(http.StatusBadRequest, models.APIError{Code: http.StatusBadRequest, Message: "invalid file type (allowed: jpg, png, gif, webp)"})
		return
	}

	// Read file content
	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{Code: http.StatusInternalServerError, Message: "failed to read file"})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("tutorials/%s%s", uuid.New().String(), ext)
	contentType := http.DetectContentType(data)

	// Upload to S3
	if err := h.s3Store.PutFile(c.Request.Context(), filename, data, contentType); err != nil {
		log.Printf("Failed to upload image to S3: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIError{Code: http.StatusInternalServerError, Message: "failed to upload image"})
		return
	}

	// Return the URL (proxied through API)
	url := fmt.Sprintf("/api/public/tutorials/images/%s", filename)
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// GetImage serves a tutorial image
// GET /api/public/tutorials/images/*path
func (h *TutorialHandler) GetImage(c *gin.Context) {
	if h.s3Store == nil {
		c.JSON(http.StatusServiceUnavailable, models.APIError{Code: http.StatusServiceUnavailable, Message: "storage not configured"})
		return
	}

	path := c.Param("path")
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	data, err := h.s3Store.GetFile(c.Request.Context(), path)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIError{Code: http.StatusNotFound, Message: "image not found"})
		return
	}

	contentType := http.DetectContentType(data)
	c.Data(http.StatusOK, contentType, data)
}
