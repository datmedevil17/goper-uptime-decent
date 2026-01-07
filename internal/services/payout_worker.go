package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/datmedevil17/gopher-uptime/internal/models"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

type PayoutWorker struct {
	db             *gorm.DB
	rabbitMQ       *amqp.Channel
	solanaClient   *rpc.Client
	platformWallet solana.PrivateKey
}

type PayoutRequest struct {
	ValidatorID string  `json:"validator_id"`
	Amount      float64 `json:"amount"`
	PublicKey   string  `json:"public_key"`
}

func NewPayoutWorker(db *gorm.DB, rabbitMQ *amqp.Channel, platformPrivateKey string) (*PayoutWorker, error) {
	// Initialize Solana client for devnet
	solanaClient := rpc.New(rpc.DevNet_RPC)

	// Parse platform private key
	privateKey, err := solana.PrivateKeyFromBase58(platformPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid platform private key: %w", err)
	}

	log.Printf("✅ Payout worker initialized with wallet: %s", privateKey.PublicKey().String())

	return &PayoutWorker{
		db:             db,
		rabbitMQ:       rabbitMQ,
		solanaClient:   solanaClient,
		platformWallet: privateKey,
	}, nil
}

// Start begins consuming from RabbitMQ
func (w *PayoutWorker) Start() error {
	// Declare queue (idempotent)
	q, err := w.rabbitMQ.QueueDeclare(
		"payout_queue", // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Set QoS - process one message at a time
	err = w.rabbitMQ.Qos(1, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Start consuming
	msgs, err := w.rabbitMQ.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack (use manual ack for reliability)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Println("💰 Payout worker started, waiting for messages...")

	// Process messages
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			w.processPayoutRequest(d)
		}
	}()

	<-forever
	return nil
}

// processPayoutRequest handles individual payout
func (w *PayoutWorker) processPayoutRequest(delivery amqp.Delivery) {
	var req PayoutRequest
	if err := json.Unmarshal(delivery.Body, &req); err != nil {
		log.Printf("❌ Error unmarshaling payout request: %v", err)
		delivery.Nack(false, false) // Don't requeue malformed messages
		return
	}

	log.Printf("💸 Processing payout for validator %s: %.2f lamports", req.ValidatorID, req.Amount)

	// Create transaction record using GORM
	txRecord := &models.PayoutTransaction{
		ID:          uuid.New().String(),
		ValidatorID: req.ValidatorID,
		Amount:      req.Amount,
		Status:      "processing",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := w.db.Create(txRecord).Error; err != nil {
		log.Printf("❌ Failed to create transaction record: %v", err)
		delivery.Nack(false, true) // Requeue
		return
	}

	// Execute Solana transfer
	signature, err := w.executeSolanaTransfer(req.PublicKey, uint64(req.Amount))
	if err != nil {
		log.Printf("❌ Solana transfer failed: %v", err)

		// Update transaction as failed using GORM
		w.db.Model(txRecord).Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": err.Error(),
			"updated_at":    time.Now(),
		})

		// Refund validator's pending balance using GORM
		w.db.Model(&models.Validator{}).
			Where("id = ?", req.ValidatorID).
			UpdateColumn("pending_payouts", gorm.Expr("pending_payouts + ?", req.Amount))

		delivery.Nack(false, false)
		return
	}

	// Poll for confirmation
	confirmed, err := w.waitForConfirmation(signature, 30*time.Second)
	if err != nil || !confirmed {
		log.Printf("❌ Transaction confirmation failed: %v", err)

		w.db.Model(txRecord).Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": "Transaction confirmation timeout",
			"tx_signature":  signature,
			"updated_at":    time.Now(),
		})

		delivery.Nack(false, false)
		return
	}

	// Update transaction as completed using GORM
	w.db.Model(txRecord).Updates(map[string]interface{}{
		"status":       "completed",
		"tx_signature": signature,
		"updated_at":   time.Now(),
	})

	log.Printf("✅ Payout completed successfully. TX: %s", signature)
	delivery.Ack(false)
}

// executeSolanaTransfer creates and sends Solana transaction
func (w *PayoutWorker) executeSolanaTransfer(recipientPublicKey string, lamports uint64) (string, error) {
	ctx := context.Background()

	// Parse recipient public key
	recipient, err := solana.PublicKeyFromBase58(recipientPublicKey)
	if err != nil {
		return "", fmt.Errorf("invalid recipient public key: %w", err)
	}

	// Get latest blockhash
	recent, err := w.solanaClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("failed to get latest blockhash: %w", err)
	}

	// Create transfer instruction
	instruction := system.NewTransferInstruction(
		lamports,
		w.platformWallet.PublicKey(),
		recipient,
	).Build()

	// Build transaction
	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		recent.Value.Blockhash,
		solana.TransactionPayer(w.platformWallet.PublicKey()),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create transaction: %w", err)
	}

	// Sign transaction
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(w.platformWallet.PublicKey()) {
			return &w.platformWallet
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	sig, err := w.solanaClient.SendTransactionWithOpts(
		ctx,
		tx,
		rpc.TransactionOpts{
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return sig.String(), nil
}

// waitForConfirmation polls for transaction confirmation
func (w *PayoutWorker) waitForConfirmation(signature string, timeout time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	sig := solana.MustSignatureFromBase58(signature)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("confirmation timeout")
		case <-ticker.C:
			status, err := w.solanaClient.GetSignatureStatuses(ctx, true, sig)
			if err != nil {
				log.Printf("⚠️  Error checking signature status: %v", err)
				continue
			}

			if len(status.Value) > 0 && status.Value[0] != nil {
				if status.Value[0].ConfirmationStatus == rpc.ConfirmationStatusFinalized {
					return true, nil
				}
				if status.Value[0].Err != nil {
					return false, fmt.Errorf("transaction failed: %v", status.Value[0].Err)
				}
			}
		}
	}
}
