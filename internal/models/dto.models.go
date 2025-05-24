package models

import "time"

// Request DTOs

type TopUpRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type PaymentRequest struct {
	Amount  float64 `json:"amount" binding:"required,gt=0"`
	Remarks string  `json:"remarks" binding:"required"`
}

type TransferRequest struct {
	TargetUser string  `json:"target_user" binding:"required,uuid"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Remarks    string  `json:"remarks" binding:"required"`
}

type UpdateProfileRequest struct {
	Firstname string `json:"first_name" binding:"required"`
	Lastname  string `json:"last_name" binding:"required"`
	Address   string `json:"address" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Response DTOs

type UserResponse struct {
	ID        string    `json:"user_id"`
	Firstname string    `json:"first_name"`
	Lastname  string    `json:"last_name"`
	Phone     string    `json:"phone_number"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_date"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TopUpResponse struct {
	ID            string    `json:"top_up_id"`
	Amount        float64   `json:"amount_top_up"`
	BalanceBefore float64   `json:"balance_before"`
	BalanceAfter  float64   `json:"balance_after"`
	CreatedAt     time.Time `json:"created_date"`
}

type PaymentResponse struct {
	ID            string    `json:"payment_id"`
	Amount        float64   `json:"amount"`
	Remarks       string    `json:"remarks"`
	BalanceBefore float64   `json:"balance_before"`
	BalanceAfter  float64   `json:"balance_after"`
	CreatedAt     time.Time `json:"created_date"`
}

type TransferResponse struct {
	ID            string    `json:"transfer_id"`
	Amount        float64   `json:"amount"`
	Remarks       string    `json:"remarks"`
	BalanceBefore float64   `json:"balance_before"`
	BalanceAfter  float64   `json:"balance_after"`
	CreatedAt     time.Time `json:"created_date"`
}

type TransactionResponse struct {
	ID              string            `json:"transaction_id"`
	Userid          string            `json:"user_id"`
	Status          TransactionStatus `json:"status"`
	TransactionType string            `json:"transaction_type"`
	Amount          float64           `json:"amount"`
	Remarks         string            `json:"remarks"`
	BalanceBefore   float64           `json:"balance_before"`
	BalanceAfter    float64           `json:"balance_after"`
	CreatedAt       time.Time         `json:"created_date"`
}

type UpdateProfileResponse struct {
	ID        string    `json:"user_id"`
	Firstname string    `json:"first_name"`
	Lastname  string    `json:"last_name"`
	Address   string    `json:"address"`
	UpdatedAt time.Time `json:"updated_date"`
}
