package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := SetupRouter()

	r.Use(gin.Recovery())

	if err := r.Run("0.0.0.0:8080"); err != nil {
		log.Fatalf("не удалось запустить сервер: %v", err)
	}
}

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	return router
}
