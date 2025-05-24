package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redha28/foomlet/internal/handlers"
	"github.com/redha28/foomlet/internal/middlewares"
	"github.com/redha28/foomlet/internal/repositories"
)

func userRoute(r *gin.RouterGroup, db *pgxpool.Pool) {
	repo := repositories.NewUserRepo(db)
	handlers := handlers.NewUserHandler(repo)

	auth := r.Group("/auth")
	{
		auth.POST("", handlers.Login)
		auth.POST("/new", handlers.Register)
		auth.POST("/refresh", handlers.RefreshToken)
	}

	profile := r.Group("/profile")
	profile.Use(middlewares.AuthMiddleware())
	{
		profile.PATCH("", handlers.UpdateProfile)
	}
}
