package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redha28/foomlet/internal/handlers"
	"github.com/redha28/foomlet/internal/middlewares"
	"github.com/redha28/foomlet/internal/repositories"
)

func transactionRoute(r *gin.RouterGroup, db *pgxpool.Pool) {
	repo := repositories.NewTransactionRepo(db)
	handlers := handlers.NewTransactionHandler(repo)

	r.POST("/topup", middlewares.AuthMiddleware(), handlers.TopUp)
	r.POST("/payments", middlewares.AuthMiddleware(), handlers.Payment)
	r.POST("/transfers", middlewares.AuthMiddleware(), handlers.Transfer)
	r.GET("/transactions", middlewares.AuthMiddleware(), handlers.GetAllTransactions)
}
