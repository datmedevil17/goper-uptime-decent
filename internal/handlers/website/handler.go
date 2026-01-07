package website

import (
	"net/http"

	"github.com/datmedevil17/gopher-uptime/internal/models"
	"github.com/datmedevil17/gopher-uptime/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// DTO for creating website
type CreateWebsiteRequest struct {
	URL string `json:"url" binding:"required,url"`
}

// CreateWebsite - POST /api/v1/website
func (h *Handler) CreateWebsite(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req CreateWebsiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Create website with GORM
	website := models.Website{
		ID:       uuid.New().String(),
		URL:      req.URL,
		UserID:   userID.(string),
		Disabled: false,
	}

	result := h.db.Create(&website)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create website")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, gin.H{
		"id":  website.ID,
		"url": website.URL,
	})
}

// GetWebsites - GET /api/v1/websites
func (h *Handler) GetWebsites(c *gin.Context) {
	userID, _ := c.Get("userID")

	var websites []models.Website

	// Use GORM Preload to eager load ticks
	result := h.db.
		Preload("Ticks", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(100)
		}).
		Where("user_id = ? AND disabled = ?", userID, false).
		Order("created_at DESC").
		Find(&websites)

	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch websites")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"websites": websites,
		"count":    len(websites),
	})
}

// GetWebsiteStatus - GET /api/v1/website/status?websiteId=xxx
func (h *Handler) GetWebsiteStatus(c *gin.Context) {
	websiteID := c.Query("websiteId")
	if websiteID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "websiteId query parameter required")
		return
	}

	userID, _ := c.Get("userID")

	var website models.Website
	result := h.db.
		Preload("Ticks", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(100)
		}).
		Where("id = ? AND user_id = ? AND disabled = ?", websiteID, userID, false).
		First(&website)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Website not found")
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Database error")
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, website)
}

// DeleteWebsite - DELETE /api/v1/website
func (h *Handler) DeleteWebsite(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		WebsiteID string `json:"websiteId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// Soft delete by setting disabled = true
	result := h.db.Model(&models.Website{}).
		Where("id = ? AND user_id = ?", req.WebsiteID, userID).
		Update("disabled", true)

	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete website")
		return
	}

	if result.RowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "Website not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Website deleted successfully",
	})
}
