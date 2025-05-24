package pkg

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func Server(router *gin.Engine) *http.Server {
	var address string = "0.0.0.0:8080"
	if port := os.Getenv("PORT"); port != "" {
		address = ":" + port
	}

	server := &http.Server{
		Addr:         address,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 15,
		Handler:      router,
	}
	return server
}
