package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rexec/rexec/internal/storage"
)

// RecordingHandler manages terminal recordings
type RecordingHandler struct {
	store           *storage.PostgresStore
	recordings      map[string]*ActiveRecording // containerID -> recording
	mu              sync.RWMutex
}

// ActiveRecording represents an in-progress recording
type ActiveRecording struct {
	ID          string
	ContainerID string
	UserID      string
	Title       string
	StartedAt   time.Time
	Events      []RecordingEvent
	mu          sync.Mutex
}

// RecordingEvent represents a single event in a recording
type RecordingEvent struct {
	Time int64  `json:"t"` // Milliseconds since start
	Type string `json:"e"` // "o" for output, "i" for input, "r" for resize
	Data string `json:"d"` // Event data
	Cols int    `json:"c,omitempty"` // For resize events
	Rows int    `json:"r,omitempty"` // For resize events
}

// RecordingMetadata represents metadata about a recording
type RecordingMetadata struct {
	Version   int       `json:"version"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Timestamp time.Time `json:"timestamp"`
	Duration  float64   `json:"duration"` // In seconds
	Title     string    `json:"title"`
	Env       map[string]string `json:"env,omitempty"`
}

// NewRecordingHandler creates a new recording handler
func NewRecordingHandler(store *storage.PostgresStore, storagePath string) *RecordingHandler {
	handler := &RecordingHandler{
		store:       store,
		recordings:  make(map[string]*ActiveRecording),
	}
	
	log.Printf("[Recording] Handler initialized (storing in database)")
	return handler
}

// StartRecording starts recording a terminal session
func (h *RecordingHandler) StartRecording(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		ContainerID string `json:"container_id" binding:"required"`
		Title       string `json:"title"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[Recording] Start request for container: %s by user: %s", req.ContainerID, userID)

	// Check if already recording this container
	h.mu.RLock()
	if _, exists := h.recordings[req.ContainerID]; exists {
		h.mu.RUnlock()
		log.Printf("[Recording] Already recording container: %s", req.ContainerID)
		c.JSON(http.StatusConflict, gin.H{"error": "already recording this terminal"})
		return
	}
	h.mu.RUnlock()

	// Set default title
	if req.Title == "" {
		req.Title = fmt.Sprintf("Recording %s", time.Now().Format("2006-01-02 15:04"))
	}

	recording := &ActiveRecording{
		ID:          uuid.New().String(),
		ContainerID: req.ContainerID,
		UserID:      userID.(string),
		Title:       req.Title,
		StartedAt:   time.Now(),
		Events:      make([]RecordingEvent, 0),
	}

	h.mu.Lock()
	h.recordings[req.ContainerID] = recording
	activeCount := len(h.recordings)
	h.mu.Unlock()

	log.Printf("[Recording] Started recording %s for container: %s (total active: %d)", recording.ID, req.ContainerID, activeCount)

	c.JSON(http.StatusOK, gin.H{
		"recording_id":   recording.ID,
		"started_at":     recording.StartedAt,
		"container_id":   req.ContainerID,
		"message":        "Recording started",
	})
}

// AddEvent adds an event to an active recording
func (h *RecordingHandler) AddEvent(containerID string, eventType string, data string, cols, rows int) {
	h.mu.RLock()
	recording, exists := h.recordings[containerID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	recording.mu.Lock()
	defer recording.mu.Unlock()

	elapsed := time.Since(recording.StartedAt).Milliseconds()

	event := RecordingEvent{
		Time: elapsed,
		Type: eventType,
		Data: data,
	}

	if eventType == "r" {
		event.Cols = cols
		event.Rows = rows
	}

	recording.Events = append(recording.Events, event)
}

// StopRecording stops and saves a recording
func (h *RecordingHandler) StopRecording(c *gin.Context) {
	userID, _ := c.Get("userID")
	containerID := c.Param("containerId")

	log.Printf("[Recording] Stop request for container: %s by user: %s", containerID, userID)

	h.mu.Lock()
	recording, exists := h.recordings[containerID]
	if !exists {
		h.mu.Unlock()
		// Log all active recordings for debugging
		h.mu.RLock()
		activeIDs := make([]string, 0, len(h.recordings))
		for id := range h.recordings {
			activeIDs = append(activeIDs, id)
		}
		h.mu.RUnlock()
		log.Printf("[Recording] No active recording for container: %s. Active recordings: %v", containerID, activeIDs)
		c.JSON(http.StatusNotFound, gin.H{"error": "no active recording for this terminal"})
		return
	}
	delete(h.recordings, containerID)
	h.mu.Unlock()

	log.Printf("[Recording] Found recording %s for container: %s, events: %d", recording.ID, containerID, len(recording.Events))

	// Verify ownership
	if recording.UserID != userID.(string) {
		log.Printf("[Recording] Unauthorized: recording user %s != request user %s", recording.UserID, userID)
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	// Calculate duration
	duration := time.Since(recording.StartedAt)

	// Generate share token
	shareToken := generateRecordingToken()

	// Convert recording to asciicast format and store in database
	recordingData, err := h.convertToAsciicast(recording)
	if err != nil {
		log.Printf("[Recording] Failed to convert recording: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process recording"})
		return
	}

	log.Printf("[Recording] Converted recording: %d bytes, %d events", len(recordingData), len(recording.Events))

	// Save to database with data
	record := &storage.RecordingRecord{
		ID:          recording.ID,
		UserID:      userID.(string),
		ContainerID: containerID,
		Title:       recording.Title,
		Duration:    duration.Milliseconds(),
		Size:        int64(len(recordingData)),
		Data:        recordingData,
		ShareToken:  shareToken,
		IsPublic:    false,
		CreatedAt:   recording.StartedAt,
	}

	if err := h.store.CreateRecording(c.Request.Context(), record); err != nil {
		log.Printf("[Recording] Failed to save recording: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save recording"})
		return
	}

	log.Printf("[Recording] Successfully saved recording %s to database", recording.ID)

	c.JSON(http.StatusOK, gin.H{
		"recording_id": recording.ID,
		"duration_ms":  duration.Milliseconds(),
		"duration":     formatDuration(duration),
		"events_count": len(recording.Events),
		"size_bytes":   len(recordingData),
		"share_token":  shareToken,
		"message":      "Recording saved",
	})
}

// GetRecordings returns all recordings for a user
func (h *RecordingHandler) GetRecordings(c *gin.Context) {
	userID, _ := c.Get("userID")

	recordings, err := h.store.GetRecordingsByUserID(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recordings"})
		return
	}

	var result []gin.H
	for _, r := range recordings {
		result = append(result, gin.H{
			"id":          r.ID,
			"title":       r.Title,
			"duration_ms": r.Duration,
			"duration":    formatDuration(time.Duration(r.Duration) * time.Millisecond),
			"size_bytes":  r.Size,
			"is_public":   r.IsPublic,
			"share_token": r.ShareToken,
			"share_url":   "/r/" + r.ShareToken,
			"created_at":  r.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"recordings": result})
}

// GetRecording returns a specific recording
func (h *RecordingHandler) GetRecording(c *gin.Context) {
	recordingID := c.Param("id")
	userID, exists := c.Get("userID")

	recording, err := h.store.GetRecordingByID(c.Request.Context(), recordingID)
	if err != nil || recording == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recording not found"})
		return
	}

	// Check authorization
	if !recording.IsPublic && (!exists || recording.UserID != userID.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          recording.ID,
		"title":       recording.Title,
		"duration_ms": recording.Duration,
		"duration":    formatDuration(time.Duration(recording.Duration) * time.Millisecond),
		"size_bytes":  recording.Size,
		"is_public":   recording.IsPublic,
		"share_url":   "/r/" + recording.ShareToken,
		"created_at":  recording.CreatedAt,
	})
}

// GetRecordingByToken returns a recording by share token (public access)
func (h *RecordingHandler) GetRecordingByToken(c *gin.Context) {
	token := c.Param("token")

	recording, err := h.store.GetRecordingByShareToken(c.Request.Context(), token)
	if err != nil || recording == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recording not found or expired"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          recording.ID,
		"title":       recording.Title,
		"duration_ms": recording.Duration,
		"duration":    formatDuration(time.Duration(recording.Duration) * time.Millisecond),
		"created_at":  recording.CreatedAt,
	})
}

// StreamRecording streams the recording data
func (h *RecordingHandler) StreamRecording(c *gin.Context) {
	recordingID := c.Param("id")
	userID, exists := c.Get("userID")

	recording, err := h.store.GetRecordingByID(c.Request.Context(), recordingID)
	if err != nil || recording == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recording not found"})
		return
	}

	// Check authorization
	if !recording.IsPublic && (!exists || recording.UserID != userID.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	// Get recording data from database
	data, err := h.store.GetRecordingData(c.Request.Context(), recordingID)
	if err != nil || data == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recording data not found"})
		return
	}

	c.Header("Content-Type", "application/x-asciicast")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.cast", recording.Title))
	
	// Decompress if gzipped
	c.Data(http.StatusOK, "application/x-asciicast", data)
}

// StreamRecordingByToken streams recording by share token
func (h *RecordingHandler) StreamRecordingByToken(c *gin.Context) {
	token := c.Param("token")

	data, err := h.store.GetRecordingDataByToken(c.Request.Context(), token)
	if err != nil || data == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recording not found"})
		return
	}

	c.Header("Content-Type", "application/x-asciicast")
	c.Data(http.StatusOK, "application/x-asciicast", data)
}

// UpdateRecording updates recording settings
func (h *RecordingHandler) UpdateRecording(c *gin.Context) {
	recordingID := c.Param("id")
	userID, _ := c.Get("userID")

	recording, err := h.store.GetRecordingByID(c.Request.Context(), recordingID)
	if err != nil || recording == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recording not found"})
		return
	}

	if recording.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	var req struct {
		IsPublic *bool  `json:"is_public"`
		Title    string `json:"title"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.IsPublic != nil {
		if err := h.store.UpdateRecordingVisibility(c.Request.Context(), recordingID, *req.IsPublic); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "recording updated"})
}

// DeleteRecording deletes a recording
func (h *RecordingHandler) DeleteRecording(c *gin.Context) {
	recordingID := c.Param("id")
	userID, _ := c.Get("userID")

	recording, err := h.store.GetRecordingByID(c.Request.Context(), recordingID)
	if err != nil || recording == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recording not found"})
		return
	}

	if recording.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	// Delete from database (data is stored in DB, no file to delete)
	if err := h.store.DeleteRecording(c.Request.Context(), recordingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "recording deleted"})
}

// IsRecording checks if a container is being recorded
func (h *RecordingHandler) IsRecording(containerID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.recordings[containerID]
	return exists
}

// GetRecordingStatus returns the status of an active recording
func (h *RecordingHandler) GetRecordingStatus(c *gin.Context) {
	containerID := c.Param("containerId")

	h.mu.RLock()
	recording, exists := h.recordings[containerID]
	h.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusOK, gin.H{"recording": false})
		return
	}

	recording.mu.Lock()
	eventsCount := len(recording.Events)
	recording.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"recording":     true,
		"recording_id":  recording.ID,
		"started_at":    recording.StartedAt,
		"duration_ms":   time.Since(recording.StartedAt).Milliseconds(),
		"events_count":  eventsCount,
	})
}

// convertToAsciicast converts recording to asciicast v2 format
func (h *RecordingHandler) convertToAsciicast(recording *ActiveRecording) ([]byte, error) {
	var buf bytes.Buffer

	// Write header (asciinema v2 format)
	header := RecordingMetadata{
		Version:   2,
		Width:     120,
		Height:    30,
		Timestamp: recording.StartedAt,
		Duration:  time.Since(recording.StartedAt).Seconds(),
		Title:     recording.Title,
	}

	headerJSON, _ := json.Marshal(header)
	buf.Write(headerJSON)
	buf.WriteByte('\n')

	// Write events
	recording.mu.Lock()
	defer recording.mu.Unlock()

	for _, event := range recording.Events {
		// asciinema format: [time, event_type, data]
		timeInSeconds := float64(event.Time) / 1000.0
		eventData := []interface{}{timeInSeconds, event.Type, event.Data}
		eventJSON, _ := json.Marshal(eventData)
		buf.Write(eventJSON)
		buf.WriteByte('\n')
	}

	return buf.Bytes(), nil
}

func generateRecordingToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return strings.ReplaceAll(base64.URLEncoding.EncodeToString(b), "=", "")
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
