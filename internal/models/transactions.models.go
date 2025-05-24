package models

import (
	"time"

	"github.com/google/uuid"
)

// Transaction Types Constants
const (
	TransactionTypeTopUp    = 1
	TransactionTypePayment  = 2
	TransactionTypeTransfer = 3
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending TransactionStatus = "PENDING"
	TransactionStatusSuccess TransactionStatus = "SUCCESS"
	TransactionStatusFailed  TransactionStatus = "FAILED"
)

// Transaction represents the transactions table
type Transaction struct {
	ID                string            `json:"id"`
	WalletID          string            `json:"wallet_id"`
	TransactionTypeID int               `json:"transaction_type_id"`
	Amount            float64           `json:"amount"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	Status            TransactionStatus `json:"status"`
	Remarks           string            `json:"remarks"`
	BalanceBefore     float64           `json:"balance_before"`
	BalanceAfter      float64           `json:"balance_after"`
}

// NewTransaction creates a new transaction with a generated UUID
func NewTransaction() *Transaction {
	return &Transaction{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    TransactionStatusSuccess,
	}
}

// TopUp represents the topup table
type TopUp struct {
	TransactionID string    `json:"transaction_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewTopUp creates a new top-up entry
func NewTopUp(transactionID string) *TopUp {
	return &TopUp{
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// Payment represents the payments table
type Payment struct {
	TransactionID string    `json:"transaction_id"`
	UserID        string    `json:"user_id"`
	Amount        int       `json:"amount"`
	Remarks       string    `json:"remarks"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewPayment creates a new payment record
func NewPayment(transactionID string) *Payment {
	return &Payment{
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// Transfer represents the transfer table
type Transfer struct {
	TransactionID string    `json:"transaction_id"`
	TargetUser    string    `json:"target_user"`
	SenderUser    string    `json:"sender_user"`
	Remarks       string    `json:"remarks"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewTransfer creates a new transfer record
func NewTransfer(transactionID string) *Transfer {
	return &Transfer{
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// TransferQueue represents a transfer that will be processed asynchronously
type TransferQueue struct {
	ID         string            `json:"id"`
	SenderUser string            `json:"sender_user"`
	TargetUser string            `json:"target_user"`
	Amount     float64           `json:"amount"`
	Remarks    string            `json:"remarks"`
	Status     TransactionStatus `json:"status"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// NewTransferQueue creates a new transfer queue entry
func NewTransferQueue() *TransferQueue {
	return &TransferQueue{
		ID:        uuid.New().String(),
		Status:    TransactionStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
