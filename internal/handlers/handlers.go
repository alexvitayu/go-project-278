package handlers

import (
	"code/internal/service"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const Short_name_length = 6

type LinkRequest struct {
	Original_url string `json:"original_url" validate:"required,url"`
	Short_name   string `json:"short_name"`
}

type Handler struct {
	service service.LinkServer
}

func NewHandler(s service.LinkServer) *Handler {
	return &Handler{service: s}
}

func (h *Handler) HomePage(c *gin.Context) {
	_ = h
	c.JSON(http.StatusOK, "URL Shortener API is running!...")
}

func (h *Handler) CreateLink(c *gin.Context) {
	request := GetRequestAndValidate(c)
	shortName := request.Short_name
	if shortName == "" {
		short, err := service.GenerateShortName(Short_name_length)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		shortName = short
	}
	link, err := h.service.CreateShortLink(c.Request.Context(), shortName, request.Original_url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, &link)
}

func (h *Handler) GetLinks(c *gin.Context) {
	links, err := h.service.GetLinks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, &links)
}

func (h *Handler) GetLinkByID(c *gin.Context) {
	id := GetIDFromRequest(c)
	link, err := h.service.GetLinkByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, &link)
}

func (h *Handler) UpdateLinkByID(c *gin.Context) {
	id := GetIDFromRequest(c)

	request := GetRequestAndValidate(c)

	shortName := request.Short_name
	if shortName == "" {
		short, err := service.GenerateShortName(Short_name_length)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		shortName = short
	}

	link, err := h.service.UpdateLinkByID(c.Request.Context(), shortName, request.Original_url, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, &link)
}

func (h *Handler) DeleteLinkByID(c *gin.Context) {
	id := GetIDFromRequest(c)
	deleted, err := h.service.DeleteLinkByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusNoContent, &deleted)
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

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": "requested API endpoint doesn't exist",
		})
	})
	return router
}
