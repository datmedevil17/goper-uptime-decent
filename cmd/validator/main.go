package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/datmedevil17/gopher-uptime/internal/config"
	"github.com/gagliardetto/solana-go"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ValidatorClient struct {
	conn        *websocket.Conn
	connMu      sync.Mutex
	keypair     solana.PrivateKey
	validatorID string
	callbacks   map[string]func(OutgoingMessage)
}

type IncomingMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type OutgoingMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type SignupData struct {
	ValidatorID string `json:"validatorId"`
	CallbackID  string `json:"callbackId"`
}

type ValidateData struct {
	URL        string `json:"url"`
	CallbackID string `json:"callbackId"`
	WebsiteID  string `json:"websiteId"`
}

func NewValidatorClient(privateKey string) (*ValidatorClient, error) {
	keypair, err := solana.PrivateKeyFromBase58(privateKey)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Validator initialized with public key: %s", keypair.PublicKey().String())

	return &ValidatorClient{
		keypair:   keypair,
		callbacks: make(map[string]func(OutgoingMessage)),
	}, nil
}

func (v *ValidatorClient) Connect(hubURL string) error {
	log.Printf("🔌 Connecting to hub: %s", hubURL)

	conn, _, err := websocket.DefaultDialer.Dial(hubURL, nil)
	if err != nil {
		return err
	}
	v.conn = conn

	log.Println("✅ Connected to hub")

	// Start listening for messages
	go v.listen()

	// Sign up with hub
	return v.signup()
}

func (v *ValidatorClient) listen() {
	for {
		var msg OutgoingMessage
		err := v.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("❌ Read error: %v", err)
			return
		}

		switch msg.Type {
		case "signup":
			v.handleSignupResponse(msg.Data)
		case "validate":
			v.handleValidateRequest(msg.Data)
		}
	}
}

func (v *ValidatorClient) signup() error {
	callbackID := uuid.New().String()
	message := "Signed message for " + callbackID + ", " + v.keypair.PublicKey().String()
	signature := v.signMessage(message)

	// Register callback for signup response
	v.callbacks[callbackID] = func(msg OutgoingMessage) {
		data := msg.Data.(map[string]interface{})
		v.validatorID = data["validatorId"].(string)
		log.Printf("✅ Validator ID received: %s", v.validatorID)
	}

	// Send signup message
	msg := IncomingMessage{
		Type: "signup",
		Data: mustMarshal(map[string]string{
			"callbackId":    callbackID,
			"ip":            "127.0.0.1",
			"publicKey":     v.keypair.PublicKey().String(),
			"signedMessage": signature,
		}),
	}

	v.connMu.Lock()
	defer v.connMu.Unlock()
	if err := v.conn.WriteJSON(msg); err != nil {
		return err
	}

	log.Println("📤 Signup request sent")
	return nil
}

func (v *ValidatorClient) handleSignupResponse(data interface{}) {
	dataMap := data.(map[string]interface{})
	callbackID := dataMap["callbackId"].(string)

	if callback, exists := v.callbacks[callbackID]; exists {
		callback(OutgoingMessage{Type: "signup", Data: data})
		delete(v.callbacks, callbackID)
	}
}

func (v *ValidatorClient) handleValidateRequest(data interface{}) {
	var validateData ValidateData
	jsonData, _ := json.Marshal(data)
	json.Unmarshal(jsonData, &validateData)

	log.Printf("📥 Validation request received: %s", validateData.URL)

	// Validate in goroutine (non-blocking)
	go v.validateWebsite(validateData)
}

func (v *ValidatorClient) validateWebsite(data ValidateData) {
	startTime := time.Now()

	// Perform HTTP GET request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(data.URL)
	latency := time.Since(startTime).Milliseconds()

	status := "Bad"
	if err == nil && resp.StatusCode == 200 {
		status = "Good"
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Sign the response
	signature := v.signMessage("Replying to " + data.CallbackID)

	// Send result back to hub
	msg := IncomingMessage{
		Type: "validate",
		Data: mustMarshal(map[string]interface{}{
			"callbackId":    data.CallbackID,
			"status":        status,
			"latency":       float64(latency),
			"validatorId":   v.validatorID,
			"websiteId":     data.WebsiteID,
			"signedMessage": signature,
		}),
	}

	v.connMu.Lock()
	if err := v.conn.WriteJSON(msg); err != nil {
		v.connMu.Unlock()
		log.Printf("❌ Failed to send validation result: %v", err)
	} else {
		v.connMu.Unlock()
		log.Printf("✅ Validation complete: %s - %s (%dms)", data.URL, status, latency)
	}
}

func (v *ValidatorClient) signMessage(message string) string {
	signature := ed25519.Sign(ed25519.PrivateKey(v.keypair), []byte(message))
	return base64.StdEncoding.EncodeToString(signature)
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

func main() {
	cfg := config.Load()

	// Get private key from environment
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("❌ PRIVATE_KEY environment variable required")
	}

	// Create validator client
	client, err := NewValidatorClient(privateKey)
	if err != nil {
		log.Fatal("❌ Failed to create validator:", err)
	}

	// Connect to hub
	if err := client.Connect(cfg.HubURL); err != nil {
		log.Fatal("❌ Failed to connect to hub:", err)
	}

	log.Println("🚀 Validator running and waiting for tasks...")

	// Wait for interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	log.Println("👋 Validator shutting down")
	client.conn.Close()
}
