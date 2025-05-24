package handlers

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redha28/foomlet/internal/middlewares"
	"github.com/redha28/foomlet/internal/models"
	"github.com/redha28/foomlet/internal/repositories"
	"github.com/redha28/foomlet/pkg"
)

type TransactionHandler struct {
	repo repositories.TransactionRepoInterface
}

func NewTransactionHandler(repo repositories.TransactionRepoInterface) *TransactionHandler {
	return &TransactionHandler{repo: repo}
}

func (h *TransactionHandler) TopUp(c *gin.Context) {
	response := models.NewResponse(c)

	// Get user ID from context
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		response.Unauthorized("User not authenticated", nil)
		return
	}

	// Parse request
	var req models.TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest("Invalid input", err.Error())
		return
	}

	// Validate amount
	if req.Amount <= 0 {
		response.BadRequest("Amount must be greater than 0", nil)
		return
	}

	// Process top-up
	result, err := h.repo.TopUp(c, userID, req.Amount)
	if err != nil {
		response.InternalServerError("Failed to process top-up", err.Error())
		return
	}

	// Return success response
	c.JSON(200, gin.H{
		"status": "SUCCESS",
		"result": result,
	})
}

func (h *TransactionHandler) Payment(c *gin.Context) {
	response := models.NewResponse(c)

	// Get user ID from context
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		response.Unauthorized("User not authenticated", nil)
		return
	}

	// Parse request
	var req models.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest("Invalid input", err.Error())
		return
	}

	// Validate amount and remarks
	if req.Amount <= 0 {
		response.BadRequest("Amount must be greater than 0", nil)
		return
	}

	if req.Remarks == "" {
		response.BadRequest("Remarks cannot be empty", nil)
		return
	}

	// Process payment
	result, err := h.repo.Payment(c, userID, req.Amount, req.Remarks)
	if err != nil {
		if strings.Contains(err.Error(), "saldo tidak cukup") {
			response.BadRequest("Saldo tidak cukup", nil)
			return
		}
		response.InternalServerError("Failed to process payment", err.Error())
		return
	}

	// Return success response
	c.JSON(200, gin.H{
		"status": "SUCCESS",
		"result": result,
	})
}

func (h *TransactionHandler) GetAllTransactions(c *gin.Context) {
	response := models.NewResponse(c)

	// Get user ID from context
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		response.Unauthorized("User not authenticated", nil)
		return
	}

	// Get transactions
	transactions, err := h.repo.GetUserTransactions(c, userID)
	if err != nil {
		response.InternalServerError("Failed to get transactions", err.Error())
		return
	}

	// Check if transactions are empty
	if len(transactions) == 0 {
		response.NotFound("No transactions found", nil)
		return
	}

	// Return success response
	c.JSON(200, gin.H{
		"status": "SUCCESS",
		"result": transactions,
	})
}

func (h *TransactionHandler) Transfer(c *gin.Context) {
	response := models.NewResponse(c)

	// Get user ID from context
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		response.Unauthorized("User not authenticated", nil)
		return
	}

	// Parse request
	var req models.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest("Invalid input", err.Error())
		return
	}

	// Validate amount and remarks
	if req.Amount <= 0 {
		response.BadRequest("Amount must be greater than 0", nil)
		return
	}

	// Don't allow transfers to self
	if userID == req.TargetUser {
		response.BadRequest("Cannot transfer to yourself", nil)
		return
	}

	// Create transfer record (this doesn't process the actual transfer yet)
	result, err := h.repo.Transfer(c, userID, req.TargetUser, req.Amount, req.Remarks)
	if err != nil {
		if errors.Is(err, repositories.ErrInsufficientBalance) {
			response.BadRequest("Saldo tidak cukup", nil)
			return
		}
		if strings.Contains(err.Error(), "user not found") {
			response.BadRequest("Recipient user not found", nil)
			return
		}
		response.InternalServerError("Failed to process transfer", err.Error())
		return
	}

	// Create transfer message
	transferMsg := pkg.TransferMessage{
		TransferID:  result.ID,
		SenderID:    userID,
		RecipientID: req.TargetUser,
		Amount:      req.Amount,
		Remarks:     req.Remarks,
	}

	// Try to publish to RabbitMQ, if not available process directly
	rmqSuccess := false
	if pkg.GlobalRabbitMQ != nil && pkg.GlobalRabbitMQ.IsReady() {
		err := pkg.GlobalRabbitMQ.PublishTransfer(transferMsg)
		if err == nil {
			rmqSuccess = true
			log.Println("Transfer queued successfully:", result.ID)
		} else {
			log.Printf("Failed to publish transfer message: %v", err)
		}
	}

	// Process transfer directly if RabbitMQ failed or isn't available
	if !rmqSuccess {
		log.Println("Processing transfer synchronously:", result.ID)

		// Create a copy of the context to prevent it from being canceled when request completes
		bgCtx := context.Background()

		// Process the transfer immediately
		go func() {
			if err := h.repo.ProcessTransfer(bgCtx, result.ID, userID, req.TargetUser, req.Amount, req.Remarks); err != nil {
				log.Printf("Failed to process transfer: %v", err)
			} else {
				log.Println("Transfer processed successfully:", result.ID)
			}
		}()
	}

	// Return success response to the client
	c.JSON(200, gin.H{
		"status": "SUCCESS",
		"result": result,
	})
}
