package documents

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"triplanner/accounts"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DocumentUploadRequest represents the request structure for uploading a document
type DocumentUploadRequest struct {
	Name        string           `form:"name" binding:"required" example:"Flight Ticket" description:"Display name of the document"`
	Category    DocumentCategory `form:"category" binding:"required" example:"tickets" description:"Document category"`
	Description *string          `form:"description" example:"Return flight ticket from NYC to Paris" description:"Optional description"`
	Notes       *string          `form:"notes" example:"Keep this handy at airport" description:"Optional user notes"`
	Tags        []string         `form:"tags" example:"flight,business-class" description:"Optional tags for organization"`
	ExpiresAt   *time.Time       `form:"expires_at" example:"2024-12-31T23:59:59Z" description:"Optional expiration date"`
	IsPublic    bool             `form:"is_public" example:"false" description:"Whether the document is publicly accessible"`
}

// DocumentUpdateRequest represents the request structure for updating a document
type DocumentUpdateRequest struct {
	Name        *string           `json:"name" example:"Flight Ticket Updated" description:"Display name of the document"`
	Category    *DocumentCategory `json:"category" example:"tickets" description:"Document category"`
	Description *string           `json:"description" example:"Updated description" description:"Optional description"`
	Notes       *string           `json:"notes" example:"Updated notes" description:"Optional user notes"`
	Tags        []string          `json:"tags" example:"flight,business-class" description:"Optional tags for organization"`
	ExpiresAt   *time.Time        `json:"expires_at" example:"2024-12-31T23:59:59Z" description:"Optional expiration date"`
	IsPublic    *bool             `json:"is_public" example:"false" description:"Whether the document is publicly accessible"`
}

// GetDocuments godoc
// @Summary Get documents for a trip plan
// @Description Retrieve all documents for a specific trip plan with optional filtering
// @Tags documents
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Param category query string false "Filter by category"
// @Param entity_type query string false "Filter by entity type"
// @Param limit query int false "Number of records to return (default: 50)"
// @Param offset query int false "Number of records to skip (default: 0)"
// @Success 200 {object} map[string]interface{} "List of documents"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip/{id}/documents [get]
func GetDocuments(c *gin.Context) {
	tripPlanIDStr := c.Param("id")

	// Validate UUID format
	tripPlanID, err := uuid.Parse(tripPlanIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip plan ID format"})
		return
	}

	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Verify trip plan access
	var hasAccess bool
	var tripPlan struct{ ID uuid.UUID }
	result := core.DB.Table("trip_plans").Select("id").Where("id = ? AND user_id = ?", tripPlanID, user.BaseModel.ID).First(&tripPlan)
	if result.Error == nil {
		hasAccess = true
	} else {
		// Check if user is a traveller in this trip
		var traveller struct{ ID uuid.UUID }
		result = core.DB.Table("travellers").Select("id").
			Where("trip_plan = ? AND user_id = ? AND is_active = ?", tripPlanID, user.BaseModel.ID, true).First(&traveller)
		hasAccess = result.Error == nil
	}

	if !hasAccess {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found or access denied"})
		return
	}

	// Get query parameters
	category := c.Query("category")
	entityType := c.Query("entity_type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Build query for documents related to this trip
	query := core.DB.Where("(entity_type = ? AND entity_id = ?) OR user_id = ?", "trip_plan", tripPlanID, user.ID)

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}

	var documents []Document
	var count int64

	// Get total count
	core.DB.Model(&Document{}).Where("(entity_type = ? AND entity_id = ?) OR user_id = ?", "trip_plan", tripPlanID, user.BaseModel.ID).Count(&count)

	// Get documents
	result = query.Order("uploaded_at DESC").Limit(limit).Offset(offset).Find(&documents)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": documents,
		"total":     count,
		"limit":     limit,
		"offset":    offset,
	})
}

// UploadDocument godoc
// @Summary Upload a document for a trip plan
// @Description Upload a new document to a trip plan
// @Tags documents
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Trip Plan ID"
// @Param file formData file true "Document file"
// @Param name formData string true "Display name of the document"
// @Param category formData string true "Document category"
// @Param description formData string false "Optional description"
// @Param notes formData string false "Optional user notes"
// @Param tags formData string false "Comma-separated tags"
// @Param expires_at formData string false "Optional expiration date (RFC3339 format)"
// @Param is_public formData boolean false "Whether the document is publicly accessible"
// @Success 201 {object} Document "Created document"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Trip plan not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /trip/{id}/documents [post]
func UploadDocument(c *gin.Context) {
	tripPlanIDStr := c.Param("id")

	// Validate UUID format
	tripPlanID, err := uuid.Parse(tripPlanIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip plan ID format"})
		return
	}

	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	// Verify access
	var hasAccess bool
	var tripPlan struct{ ID uuid.UUID }
	result := core.DB.Table("trip_plans").Select("id").Where("id = ? AND user_id = ?", tripPlanID, user.BaseModel.ID).First(&tripPlan)
	if result.Error == nil {
		hasAccess = true
	} else {
		var traveller struct{ ID uuid.UUID }
		result = core.DB.Table("travellers").Select("id").
			Where("trip_plan = ? AND user_id = ? AND is_active = ?", tripPlanID, user.BaseModel.ID, true).First(&traveller)
		hasAccess = result.Error == nil
	}

	if !hasAccess {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip plan not found or access denied"})
		return
	}

	// Parse multipart form
	err = c.Request.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form"})
		return
	}

	// Get uploaded file
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	defer func() { _ = file.Close() }()

	// Validate file size (max 50MB)
	if fileHeader.Size > 50*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 50MB limit"})
		return
	}

	// Get form data
	name := c.PostForm("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	category := c.PostForm("category")
	if !IsValidCategory(category) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	description := c.PostForm("description")
	notes := c.PostForm("notes")
	tagsStr := c.PostForm("tags")
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}

	isPublic := c.PostForm("is_public") == "true"

	var expiresAt *time.Time
	expiresAtStr := c.PostForm("expires_at")
	if expiresAtStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, expiresAtStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expires_at format, use RFC3339"})
			return
		}
		expiresAt = &parsedTime
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := "uploads/documents"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	fileExt := filepath.Ext(fileHeader.Filename)
	fileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), fileExt)
	filePath := filepath.Join(uploadsDir, fileName)

	// Save file to disk
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer func() { _ = dst.Close() }()

	_, err = io.Copy(dst, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Create document record
	document := Document{
		Name:            name,
		OriginalName:    fileHeader.Filename,
		StorageProvider: StorageProviderLocal,
		StoragePath:     filePath,
		FileSize:        fileHeader.Size,
		ContentType:     fileHeader.Header.Get("Content-Type"),
		Category:        DocumentCategory(category),
		Description:     &description,
		Notes:           &notes,
		Tags:            tags,
		EntityType:      stringPtr("trip_plan"),
		EntityID:        &tripPlanID,
		UserID:          user.ID,
		UploadedAt:      time.Now(),
		ExpiresAt:       expiresAt,
		IsPublic:        isPublic,
	}

	if description == "" {
		document.Description = nil
	}
	if notes == "" {
		document.Notes = nil
	}

	result = core.DB.Create(&document)
	if result.Error != nil {
		// Clean up file if database insert fails
		_ = os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"document": document})
}

// GetDocument godoc
// @Summary Get a specific document
// @Description Retrieve a document by ID
// @Tags documents
// @Produce json
// @Param id path string true "Document ID"
// @Success 200 {object} Document "Document details"
// @Failure 404 {object} map[string]string "Document not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /documents/{id} [get]
func GetDocument(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var document Document
	// Check if user has access to this document
	result := core.DB.Where("id = ? AND (user_id = ? OR is_public = ?)", id, user.BaseModel.ID, true).First(&document)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"document": document})
}

// UpdateDocument godoc
// @Summary Update a document
// @Description Update an existing document
// @Tags documents
// @Accept json
// @Produce json
// @Param id path string true "Document ID"
// @Param document body DocumentUpdateRequest true "Updated document data"
// @Success 200 {object} Document "Updated document"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Document not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /documents/{id} [put]
func UpdateDocument(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var document Document
	// Verify access - user must own the document
	result := core.DB.Where("id = ? AND user_id = ?", id, user.BaseModel.ID).First(&document)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found or access denied"})
		return
	}

	var updateReq DocumentUpdateRequest
	if err := c.BindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	updates := make(map[string]interface{})
	if updateReq.Name != nil {
		updates["name"] = *updateReq.Name
	}
	if updateReq.Category != nil {
		if !IsValidCategory(string(*updateReq.Category)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
			return
		}
		updates["category"] = *updateReq.Category
	}
	if updateReq.Description != nil {
		updates["description"] = *updateReq.Description
	}
	if updateReq.Notes != nil {
		updates["notes"] = *updateReq.Notes
	}
	if updateReq.Tags != nil {
		updates["tags"] = updateReq.Tags
	}
	if updateReq.ExpiresAt != nil {
		updates["expires_at"] = *updateReq.ExpiresAt
	}
	if updateReq.IsPublic != nil {
		updates["is_public"] = *updateReq.IsPublic
	}

	result = core.DB.Model(&document).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Reload document
	core.DB.First(&document, document.ID)

	c.JSON(http.StatusOK, gin.H{"document": document})
}

// DeleteDocument godoc
// @Summary Delete a document
// @Description Delete a document and its associated file
// @Tags documents
// @Param id path string true "Document ID"
// @Success 204 "No content"
// @Failure 404 {object} map[string]string "Document not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /documents/{id} [delete]
func DeleteDocument(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var document Document
	// Verify access - user must own the document
	result := core.DB.Where("id = ? AND user_id = ?", id, user.BaseModel.ID).First(&document)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found or access denied"})
		return
	}

	// Delete file from storage if it's local storage
	if document.StorageProvider == StorageProviderLocal {
		if err := os.Remove(document.StoragePath); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to delete file %s: %v\n", document.StoragePath, err)
		}
	}

	// Delete document record
	result = core.DB.Delete(&document)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// DownloadDocument godoc
// @Summary Download a document
// @Description Download the actual file content of a document
// @Tags documents
// @Produce application/octet-stream
// @Param id path string true "Document ID"
// @Success 200 {file} file "Document file"
// @Failure 404 {object} map[string]string "Document not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /documents/{id}/download [get]
func DownloadDocument(c *gin.Context) {
	id := c.Param("id")
	currentUser, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	user := currentUser.(accounts.User)

	var document Document
	// Check if user has access to this document
	result := core.DB.Where("id = ? AND (user_id = ? OR is_public = ?)", id, user.BaseModel.ID, true).First(&document)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	// For local storage, serve the file directly
	if document.StorageProvider == StorageProviderLocal {
		if _, err := os.Stat(document.StoragePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found on disk"})
			return
		}

		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", document.OriginalName))
		c.Header("Content-Type", document.ContentType)
		c.File(document.StoragePath)
		return
	}

	// For other storage providers, implement download logic here
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Download from external storage not implemented yet"})
}

// Helper function to get string pointer
func stringPtr(s string) *string {
	return &s
}
