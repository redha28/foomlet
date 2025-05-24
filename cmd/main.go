package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/redha28/foomlet/internal/config"
	"github.com/redha28/foomlet/internal/models"
	"github.com/redha28/foomlet/internal/repositories"
	"github.com/redha28/foomlet/internal/routes"
	"github.com/redha28/foomlet/pkg"
)

var dbpool *pgxpool.Pool

func main() {
	// Initialize configuration
	if err := config.Initialize(); err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	// Set up channel to listen for OS signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pg, err := pkg.Posql()
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	defer pg.Close()
	log.Println("DB connected successfully")

	// Initialize RabbitMQ
	if err := pkg.InitRabbitMQ(); err != nil {
		log.Println("WARNING: RabbitMQ initialization failed:", err)
		log.Println("Transfers will be processed synchronously")
	} else {
		log.Println("RabbitMQ connected successfully")
		defer pkg.GlobalRabbitMQ.Close()

		// Start transfer worker in a goroutine if RabbitMQ is connected
		if pkg.GlobalRabbitMQ.IsReady() {
			// Create transaction repository for worker
			transactionRepo := repositories.NewTransactionRepo(pg)
			go runTransferWorker(ctx, transactionRepo)
		}
	}

	router := routes.InitRouter(pg)

	router.GET("/ping", func(c *gin.Context) {
		responder := models.NewResponse(c)
		reqCtx := c.Request.Context()
		if err := pg.Ping(reqCtx); err != nil {
			responder.InternalServerError("Database not responding", err.Error())
			return
		}
		responder.Success("pong", nil)
	})

	// Create HTTP server
	srv := pkg.Server(router)

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("HTTP server: %v", err)
		}
	}()
	log.Println("Server started on", srv.Addr)

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server...")

	// Cancel context to signal worker to stop
	cancel()

	// Create a deadline for server shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server gracefully
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited properly")
}

// runTransferWorker processes transfers from RabbitMQ queue
func runTransferWorker(ctx context.Context, repo repositories.TransactionRepoInterface) {
	log.Println("Starting transfer worker...")

	// Get messages from RabbitMQ
	msgs, err := pkg.GlobalRabbitMQ.ConsumeTransfers()
	if err != nil {
		log.Printf("Failed to consume transfers: %v", err)
		return
	}

	log.Println("Transfer worker started, processing messages from queue")

	for {
		select {
		case <-ctx.Done():
			log.Println("Transfer worker stopping due to context cancellation")
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Transfer worker stopping - channel closed")
				return
			}

			log.Printf("Received transfer message: %s", msg.Body)

			var transferMsg pkg.TransferMessage
			if err := json.Unmarshal(msg.Body, &transferMsg); err != nil {
				log.Printf("Error unmarshalling message: %v", err)
				msg.Reject(false) // Don't requeue malformed messages
				continue
			}

			log.Printf("Processing transfer: %s from %s to %s for %.2f",
				transferMsg.TransferID,
				transferMsg.SenderID,
				transferMsg.RecipientID,
				transferMsg.Amount)

			// Process the transfer
			processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			err := repo.ProcessTransfer(
				processCtx,
				transferMsg.TransferID,
				transferMsg.SenderID,
				transferMsg.RecipientID,
				transferMsg.Amount,
				transferMsg.Remarks,
			)
			cancel()

			if err != nil {
				log.Printf("Error processing transfer %s: %v", transferMsg.TransferID, err)

				// Determine if the error is retryable
				if isRetryableError(err) {
					log.Printf("Retrying transfer %s", transferMsg.TransferID)
					msg.Reject(true) // Requeue for retry
				} else {
					log.Printf("Permanent failure for transfer %s", transferMsg.TransferID)
					msg.Reject(false) // Don't requeue
				}
				continue
			}

			// Acknowledge successful processing
			msg.Ack(false)
			log.Printf("Successfully processed transfer %s", transferMsg.TransferID)
		}
	}
}

// Helper function to determine if an error is temporary and should be retried
func isRetryableError(err error) bool {
	return strings.Contains(err.Error(), "connection reset") ||
		strings.Contains(err.Error(), "temporarily unavailable")
}
