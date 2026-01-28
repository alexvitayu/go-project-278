package handlers

import (
	"code/internal/service"
	"context"
	"database/sql"
	"errors"
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
	Default_Limit     = 10
	Default_Offset    = 0
	Max_Limit         = 30
)

type LinkRequest struct {
	Original_url string `json:"original_url" validate:"required,url"`
	Short_name   string `json:"short_name"`
}

type Handler struct {
	linkService  service.LinkServer
	visitService service.VisitServer
}

func NewHandler(ls service.LinkServer, vs service.VisitServer) *Handler {
	return &Handler{linkService: ls, visitService: vs}
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
	link, err := h.linkService.CreateShortLink(c.Request.Context(), shortName, request.Original_url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, &link)
}

func (h *Handler) GetLinks(c *gin.Context) {
	handleGetWithRange[*service.Link](c, h.linkService.GetLinks, "links")
}

func (h *Handler) GetLinkByID(c *gin.Context) {
	id := GetIDFromRequest(c)
	link, err := h.linkService.GetLinkByID(c.Request.Context(), id)
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

	link, err := h.linkService.UpdateLinkByID(c.Request.Context(), shortName, request.Original_url, id)
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
	deleted, err := h.linkService.DeleteLinkByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusNoContent, &deleted)
}

func (h *Handler) RedirectByShortName(c *gin.Context) {
	shortName := c.Param("code")
	if shortName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "point out short_name",
		})
		return
	}
	link, err := h.linkService.GetOriginalURLByShortName(c.Request.Context(), shortName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "link not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")
	status, err := SaveConvertToInt32(http.StatusFound)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	err = h.visitService.CreateVisit(c.Request.Context(), link.ID, c.ClientIP(), userAgent, referer, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.Redirect(http.StatusFound, link.OriginalUrl)
}

func (h *Handler) GetVisits(c *gin.Context) {
	handleGetWithRange[*service.Visit](c, h.visitService.GetVisits, "link_visits")
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

func ParseAndValidateQuery(c *gin.Context) (int32, int32, error) {
	query := c.Query("range")
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

func handleGetWithRange[T any](
	c *gin.Context,
	getFunc func(ctx context.Context, limit, offset int32) ([]T, int64, error),
	resourceName string,
) {
	limit, offset, err := ParseAndValidateQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, total, err := getFunc(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Range",
		fmt.Sprintf("%s %d-%d/%d", resourceName, offset, limit+offset, total))
	c.JSON(http.StatusOK, items)
}
