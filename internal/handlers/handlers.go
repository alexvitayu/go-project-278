package handlers

import (
	"code/internal/service"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	Short_name_length = 6
	Default_Limit     = 5
	Default_Offset    = 0
	Max_Limit         = 30
)

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
	query := c.Query("range")

	limit, offset, err := ParseAndValidateQuery(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	links, total, err := h.service.GetLinks(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", offset, limit+offset, total))
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

func ParseAndValidateQuery(query string) (int32, int32, error) {
	//Установим значения по умолчанию
	if query == "" {
		return Default_Limit, Default_Offset, nil
	}
	trimmed := strings.Trim(query, "[]")
	parts := strings.Split(trimmed, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid len of range")
	}
	begin, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to convert to int: %w", err)
	}
	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to convert to int: %w", err)
	}
	if begin < 0 {
		return 0, 0, fmt.Errorf("begin can't be less than zero")
	}
	if end <= begin {
		return 0, 0, fmt.Errorf("end must be greater than begin")
	}
	limit, err := SaveConvertToInt32(end - begin)
	if err != nil {
		return 0, 0, fmt.Errorf("saveConvertToInt32: %w", err)
	}
	if limit > Max_Limit {
		limit = Max_Limit
	}
	offset, err := SaveConvertToInt32(begin)
	if err != nil {
		return 0, 0, fmt.Errorf("saveConvertToInt32: %w", err)
	}
	return limit, offset, nil
}

func SaveConvertToInt32(n int) (int32, error) {
	if n > math.MaxInt32 || n < math.MinInt32 {
		return 0, fmt.Errorf("integer overflow: %d", n)
	}
	return int32(n), nil
}
