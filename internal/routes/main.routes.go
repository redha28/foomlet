package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitRouter(pg *pgxpool.Pool) *gin.Engine {
	router := gin.Default()
	rg := router.Group("/api")
	userRoute(rg, pg)
	transactionRoute(rg, pg)
	return router
}
