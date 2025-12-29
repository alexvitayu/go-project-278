package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

func main() {
	router := SetupRouter()

	router.Use(gin.Recovery())

	if err := router.Run("0.0.0.0:8080"); err != nil {
		log.Fatalf("не удалось запустить сервер: %v", err)
	}
}

func SetupRouter() *gin.Engine {
	// To initialize Sentry's handler, you need to initialize Sentry itself beforehand
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
		// Enable structured logs to Sentry
		EnableLogs: true,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	// Create my app
	router := gin.Default()

	// Once it's done, you can attach the handler as one of your middleware
	router.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))

	// Set up route
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.GET("/check_panic", func(c *gin.Context) {
		_ = c
		panic("I check panic processing in sentry")
	})

	return router
}
