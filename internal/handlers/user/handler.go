package user

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/datmedevil17/gopher-uptime/internal/config"
	"github.com/datmedevil17/gopher-uptime/internal/models"
	"github.com/datmedevil17/gopher-uptime/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Handler struct {
	db       *gorm.DB
	rabbitMQ *amqp.Channel
	cfg      *config.Config
}

func NewHandler(db *gorm.DB, rabbitMQ *amqp.Channel, cfg *config.Config) *Handler {
	return &Handler{
		db:       db,
		rabbitMQ: rabbitMQ,
		cfg:      cfg,
	}
}

type PayoutRequest struct {
	ValidatorID string  `json:"validator_id"`
	Amount      float64 `json:"amount"`
	PublicKey   string  `json:"public_key"`
}

// RequestPayout - POST /api/v1/payout/:validatorId
func (h *Handler) RequestPayout(c *gin.Context) {
	validatorID := c.Param("validatorId")

	// Start transaction with GORM
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Lock validator row with GORM
	var validator models.Validator
	result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", validatorID).
		First(&validator)

	if result.Error != nil {
		tx.Rollback()
		if result.Error == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Validator not found")
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Database error")
		}
		return
	}

	// Check pending balance
	// Check pending balance
	if validator.PendingPayouts <= 0 {
		tx.Rollback()
		utils.SuccessResponse(c, http.StatusOK, gin.H{
			"status":  "cleared",
			"message": "all payment cleared",
			"amount":  0,
		})
		return
	}

	// Create payout request for RabbitMQ
	payoutReq := PayoutRequest{
		ValidatorID: validator.ID,
		Amount:      validator.PendingPayouts,
		PublicKey:   validator.PublicKey,
	}

	payoutJSON, err := json.Marshal(payoutReq)
	if err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to serialize request")
		return
	}

	// Publish to RabbitMQ
	err = h.rabbitMQ.Publish(
		"",             // exchange
		"payout_queue", // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payoutJSON,
			Timestamp:   time.Now(),
		},
	)

	if err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to queue payout")
		return
	}

	// Reset pending payouts using GORM
	result = tx.Model(&validator).Update("pending_payouts", 0)
	if result.Error != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update balance")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"status":  "queued",
		"message": "Payout request queued for processing",
		"amount":  payoutReq.Amount,
	})
}

// GetValidatorBalance - GET /api/v1/validator/:validatorId/balance
func (h *Handler) GetValidatorBalance(c *gin.Context) {
	validatorID := c.Param("validatorId")

	var validator models.Validator
	result := h.db.Where("id = ?", validatorID).First(&validator)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Validator not found")
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Database error")
		}
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"validator_id":        validator.ID,
		"public_key":          validator.PublicKey,
		"pending_payouts":     validator.PendingPayouts,
		"pending_payouts_sol": validator.PendingPayouts / 1e9,
	})
}

type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Signup - POST /api/v1/auth/signup
func (h *Handler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user exists
	var existingUser models.User
	if result := h.db.Where("email = ?", req.Email).First(&existingUser); result.Error == nil {
		utils.ErrorResponse(c, http.StatusConflict, "User already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create user
	user := models.User{
		ID:       uuid.New().String(),
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if result := h.db.Create(&user); result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate JWT
	token, err := utils.GenerateJWT(user.ID, h.cfg.JWTSecret)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
		},
	})
}

// Login - POST /api/v1/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Find user
	var user models.User
	if result := h.db.Where("email = ?", req.Email).First(&user); result.Error != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT
	token, err := utils.GenerateJWT(user.ID, h.cfg.JWTSecret)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
		},
	})
}
