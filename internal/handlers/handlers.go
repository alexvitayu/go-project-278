package handlers

import (
	"code/internal/service"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type LinkRequest struct {
	Original_url string `json:"original_url" validate:"required,url"`
	Short_name   string `json:"short_name"`
}

const short_name_length = 6

func SetupRouter(ctx context.Context, s *service.LinkService) *gin.Engine {
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

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": "requested API endpoint doesn't exist",
		})
	})

	// POST /api/links
	router.POST("/api/links", func(c *gin.Context) {
		request := GetRequestAndValidate(c)
		shortName := request.Short_name
		if shortName == "" {
			shortName = service.GenerateShortName(short_name_length)
		}
		link, err := s.CreateShortLink(ctx, shortName, request.Original_url)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, &link)
		return
	})

	// GET /api/links - возвращает список всех ссылок
	router.GET("/api/links", func(c *gin.Context) {
		links, err := s.GetLinks(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, &links)
		return
	})

	// GET /api/links/:id
	router.GET("/api/links/:id", func(c *gin.Context) {
		id := GetIDFromRequest(c)
		link, err := s.GetLinkByID(ctx, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, &link)
		return
	})

	// PUT /api/links/:id
	router.PUT("/api/links/:id", func(c *gin.Context) {
		id := GetIDFromRequest(c)

		request := GetRequestAndValidate(c)

		shortName := request.Short_name
		if shortName == "" {
			shortName = service.GenerateShortName(short_name_length)
		}

		link, err := s.UpdateLinkByID(ctx, shortName, request.Original_url, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, &link)
	})

	// DELETE /api/links/:id
	router.DELETE("/api/links/:id", func(c *gin.Context) {
		id := GetIDFromRequest(c)
		n, err := s.DeleteLinkByID(ctx, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		c.JSON(http.StatusNoContent, gin.H{
			"deleted": n,
		})
	})
	return router
}

func GetRequestAndValidate(c *gin.Context) *LinkRequest {
	var request LinkRequest

	if err := c.ShouldBindJSON(&request); err != nil { // Ошибка включает и ошибки парсинга, и ошибки валидации
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": err.Error(),
		})
		return &LinkRequest{}
	}

	validate := validator.New()
	if err := validate.Struct(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return &LinkRequest{}
	}
	return &request
}

func GetIDFromRequest(c *gin.Context) int64 {
	id := c.Param("id")
	intID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return 0
	}
	return int64(intID)
}
