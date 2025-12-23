package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rexec/rexec/internal/models"
	"github.com/rexec/rexec/internal/storage"
)

// TutorialHandler handles tutorial API endpoints
type TutorialHandler struct {
	store *storage.PostgresStore
}

// NewTutorialHandler creates a new TutorialHandler
func NewTutorialHandler(store *storage.PostgresStore) *TutorialHandler {
	return &TutorialHandler{
		store: store,
	}
}

// CreateTutorialRequest represents the request to create a tutorial
type CreateTutorialRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	VideoURL    string `json:"video_url" binding:"required"`
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
		VideoURL:    req.VideoURL,
		Thumbnail:   req.Thumbnail,
		Duration:    req.Duration,
		Category:    req.Category,
		Order:       req.Order,
		IsPublished: req.IsPublished,
		CreatedAt:   now,
		UpdatedAt:   now,
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
