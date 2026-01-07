package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
	
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/datmedevil17/gopher-uptime/internal/config"
	"github.com/datmedevil17/gopher-uptime/internal/database"
	"github.com/datmedevil17/gopher-uptime/internal/models"
	"gorm.io/gorm"
)

const COST_PER_VALIDATION = 100 // lamports

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type ValidatorConnection struct {
	ValidatorID string
	PublicKey   string
	Conn        *websocket.Conn
}

type Hub struct {
	db         *gorm.DB
	validators map[string]*ValidatorConnection
	mu         sync.RWMutex
	callbacks  map[string]func(IncomingMessage)
	callbackMu sync.RWMutex
}

type IncomingMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type SignupIncoming struct {
	IP            string `json:"ip"`
	PublicKey     string `json:"publicKey"`
	SignedMessage string `json:"signedMessage"`
	CallbackID    string `json:"callbackId"`
}

type ValidateIncoming struct {
	CallbackID    string  `json:"callbackId"`
	Status        string  `json:"status"`
	Latency       float64 `json:"latency"`
	ValidatorID   string  `json:"validatorId"`
	WebsiteID     string  `json:"websiteId"`
	SignedMessage string  `json:"signedMessage"`
}

type OutgoingMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewHub(db *gorm.DB) *Hub {
	return &Hub{
		db:         db,
		validators: make(map[string]*ValidatorConnection),
		callbacks:  make(map[string]func(IncomingMessage)),
	}
}

func (h *Hub) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Println("🔌 New WebSocket connection")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("❌ Read error: %v", err)
			h.removeValidator(conn)
			break
		}

		var msg IncomingMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("❌ Unmarshal error: %v", err)
			continue
		}

		switch msg.Type {
		case "signup":
			h.handleSignup(conn, msg.Data)
		case "validate":
			h.handleValidate(msg.Data)
		}
	}
}

func (h *Hub) handleSignup(conn *websocket.Conn, data json.RawMessage) {
	var signup SignupIncoming
	if err := json.Unmarshal(data, &signup); err != nil {
		log.Printf("❌ Signup unmarshal error: %v", err)
		return
	}

	// TODO: Verify signature using nacl (skipped for brevity)
	// verified := verifyMessage(...)

	var validator models.Validator

	// Find or create validator using GORM
	result := h.db.Where("public_key = ?", signup.PublicKey).First(&validator)

	if result.Error == gorm.ErrRecordNotFound {
		// Create new validator
		validator = models.Validator{
			ID:        uuid.New().String(),
			PublicKey: signup.PublicKey,
			Location:  "unknown",
			IP:        signup.IP,
		}

		if err := h.db.Create(&validator).Error; err != nil {
			log.Printf("❌ Failed to create validator: %v", err)
			return
		}
		log.Printf("✅ New validator created: %s", validator.ID)
	} else if result.Error != nil {
		log.Printf("❌ Database error: %v", result.Error)
		return
	}

	// Store validator connection
	h.mu.Lock()
	h.validators[validator.ID] = &ValidatorConnection{
		ValidatorID: validator.ID,
		PublicKey:   validator.PublicKey,
		Conn:        conn,
	}
	h.mu.Unlock()

	// Send response
	response := OutgoingMessage{
		Type: "signup",
		Data: map[string]string{
			"validatorId": validator.ID,
			"callbackId":  signup.CallbackID,
		},
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Printf("❌ Failed to send signup response: %v", err)
	} else {
		log.Printf("✅ Validator registered: %s (%s)", validator.ID, validator.PublicKey)
	}
}

func (h *Hub) handleValidate(data json.RawMessage) {
	var validate ValidateIncoming
	if err := json.Unmarshal(data, &validate); err != nil {
		log.Printf("❌ Validate unmarshal error: %v", err)
		return
	}

	// Execute callback
	h.callbackMu.RLock()
	callback, exists := h.callbacks[validate.CallbackID]
	h.callbackMu.RUnlock()

	if exists {
		var msg IncomingMessage
		msg.Type = "validate"
		msg.Data = data
		callback(msg)

		// Remove callback after execution
		h.callbackMu.Lock()
		delete(h.callbacks, validate.CallbackID)
		h.callbackMu.Unlock()
	}
}

func (h *Hub) removeValidator(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for id, validator := range h.validators {
		if validator.Conn == conn {
			delete(h.validators, id)
			log.Printf("🔌 Validator disconnected: %s", id)
			break
		}
	}
}

func (h *Hub) startMonitoring() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	log.Println("🔄 Starting monitoring loop (every 60 seconds)")

	for range ticker.C {
		var websites []models.Website

		// Fetch all active websites using GORM
		if err := h.db.Where("disabled = ?", false).Find(&websites).Error; err != nil {
			log.Printf("❌ Failed to fetch websites: %v", err)
			continue
		}

		if len(websites) == 0 {
			log.Println("⚠️  No websites to monitor")
			continue
		}

		// Get current validators
		h.mu.RLock()
		validators := make([]*ValidatorConnection, 0, len(h.validators))
		for _, v := range h.validators {
			validators = append(validators, v)
		}
		h.mu.RUnlock()

		if len(validators) == 0 {
			log.Println("⚠️  No validators connected")
			continue
		}

		log.Printf("📊 Monitoring %d websites with %d validators", len(websites), len(validators))

		// Send validation tasks
		for _, website := range websites {
			for _, validator := range validators {
				callbackID := uuid.New().String()

				// Register callback
				h.callbackMu.Lock()
				h.callbacks[callbackID] = h.createValidateCallback(website.ID, validator.PublicKey)
				h.callbackMu.Unlock()

				// Send validation request
				msg := OutgoingMessage{
					Type: "validate",
					Data: map[string]interface{}{
						"url":        website.URL,
						"callbackId": callbackID,
						"websiteId":  website.ID,
					},
				}

				if err := validator.Conn.WriteJSON(msg); err != nil {
					log.Printf("❌ Failed to send to validator %s: %v", validator.ValidatorID, err)
				} else {
					log.Printf("📤 Sent validation task: %s to %s", website.URL, validator.ValidatorID)
				}
			}
		}
	}
}

func (h *Hub) createValidateCallback(websiteID, validatorPublicKey string) func(IncomingMessage) {
	return func(msg IncomingMessage) {
		var validate ValidateIncoming
		if err := json.Unmarshal(msg.Data, &validate); err != nil {
			log.Printf("❌ Callback unmarshal error: %v", err)
			return
		}

		// TODO: Verify signature

		// Use GORM transaction
		tx := h.db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// Create tick
		tick := models.WebsiteTick{
			ID:          uuid.New().String(),
			WebsiteID:   websiteID,
			ValidatorID: validate.ValidatorID,
			Status:      validate.Status,
			Latency:     validate.Latency,
			CreatedAt:   time.Now(),
		}

		if err := tx.Create(&tick).Error; err != nil {
			tx.Rollback()
			log.Printf("❌ Failed to create tick: %v", err)
			return
		}

		// Update validator pending payouts
		if err := tx.Model(&models.Validator{}).
			Where("id = ?", validate.ValidatorID).
			UpdateColumn("pending_payouts", gorm.Expr("pending_payouts + ?", COST_PER_VALIDATION)).
			Error; err != nil {
			tx.Rollback()
			log.Printf("❌ Failed to update payouts: %v", err)
			return
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			log.Printf("❌ Failed to commit: %v", err)
			return
		}

		log.Printf("✅ Tick recorded: %s - %s (%s)", websiteID, validate.Status, validate.ValidatorID)
	}
}

func main() {
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("❌ Database connection failed:", err)
	}

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatal("❌ Migration failed:", err)
	}

	// Create hub
	hub := NewHub(db)

	// Setup HTTP handler
	http.HandleFunc("/", hub.handleWebSocket)

	// Start monitoring in background
	go hub.startMonitoring()

	// Start server
	port := "8081"
	log.Printf("🚀 Hub server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}