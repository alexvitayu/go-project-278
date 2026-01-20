package handlers_test

import (
	"bytes"
	"code/internal/handlers"
	"code/internal/handlers/mocks"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"code/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setUpRouter(t *testing.T) (*gin.Engine, *mocks.MockLinkService) {
	t.Helper()

	// Создадим моковое хранилище и передадим хендлерам
	mock := new(mocks.MockLinkService)
	handler := handlers.NewHandler(mock)

	// Создадим тестовый роутер
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Зададим тестовые маршруты
	router.POST("/api/links", handler.CreateLink)
	router.GET("/api/links", handler.GetLinks)
	router.GET("/api/links/:id", handler.GetLinkByID)
	router.PUT("/api/links/:id", handler.UpdateLinkByID)
	router.DELETE("/api/links/:id", handler.DeleteLinkByID)

	return router, mock
}

func TestHandler_CreateLink(t *testing.T) {
	t.Parallel()
	// Arrange
	router, m := setUpRouter(t)

	requestParams := map[string]string{
		"original_url": "https://example.com/very/long/url",
		"short_name":   "test123",
	}
	jsonBody, _ := json.Marshal(&requestParams)

	expectedShortUrl := "http://localhost:8080/test123"

	expectedLink := &service.Link{
		ID:          1,
		OriginalUrl: "https://example.com/very/long/url",
		ShortName:   "test123",
		ShortUrl:    expectedShortUrl,
	}

	// Записываем в моковое хранилище, что хотим передать и что ожидаем
	m.On("CreateShortLink", mock.Anything, "test123", "https://example.com/very/long/url").
		Return(expectedLink, nil).Once()

	// Act
	req := httptest.NewRequest("POST", "/api/links", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp *service.Link
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, int64(1), resp.ID)
	assert.Equal(t, expectedShortUrl, resp.ShortUrl)
	m.AssertExpectations(t)
}

func TestHandler_GetLinks(t *testing.T) {
	t.Parallel()
	//Arrange
	router, m := setUpRouter(t)

	expectedShortUrl1 := "http://localhost:8080/test1"
	expectedShortUrl2 := "http://localhost:8080/test2"

	m.On("GetLinks", mock.Anything).Return([]*service.Link{
		{ID: 1, OriginalUrl: "http://test1@gmail.com/long1", ShortName: "test1", ShortUrl: "http://localhost:8080/test1"},
		{ID: 2, OriginalUrl: "http://test2@gmail.com/long2", ShortName: "test2", ShortUrl: "http://localhost:8080/test2"},
	}, nil)

	//Act
	req := httptest.NewRequest("GET", "/api/links", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	//Assert

	var response []service.Link
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response, 2)

	assert.Equal(t, expectedShortUrl1, response[0].ShortUrl)
	assert.Equal(t, expectedShortUrl2, response[1].ShortUrl)
	m.AssertExpectations(t)
}

func TestHandler_GetLinkByID(t *testing.T) {
	t.Parallel()
	router, m := setUpRouter(t)

	linkID := int64(4)
	expectedShortUrl := "http://localhost:8080/test1"

	m.On("GetLinkByID", mock.Anything, linkID).Return(&service.Link{
		ID:          4,
		OriginalUrl: "https://example@mail.ru",
		ShortName:   "test1",
		ShortUrl:    "http://localhost:8080/test1",
	}, nil).Once()

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/links/%d", linkID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response service.Link
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedShortUrl, response.ShortUrl)
	m.AssertExpectations(t)
}

func TestHandler_UpdateLinkByID(t *testing.T) {
	t.Parallel()
	router, m := setUpRouter(t)

	linkID := int64(3)
	expectedShortUrl := "http://localhost:8080/updated"

	requestParams := map[string]string{
		"original_url": "https://example.com/very/long/url",
		"short_name":   "updated",
	}
	jsonBody, _ := json.Marshal(&requestParams)

	m.On("UpdateLinkByID", mock.Anything, "updated", "https://example.com/very/long/url", linkID).
		Return(&service.Link{
			ID:          3,
			OriginalUrl: "https://example.com/very/long/url",
			ShortName:   "updated",
			ShortUrl:    expectedShortUrl,
		}, nil).Once()

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/links/%d", linkID), bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response service.Link
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedShortUrl, response.ShortUrl)
	m.AssertExpectations(t)
}

func TestHandler_DeleteLinkByID(t *testing.T) {
	t.Parallel()
	router, m := setUpRouter(t)

	linkID := int64(5)
	expectedCode := http.StatusNoContent

	m.On("DeleteLinkByID", mock.Anything, linkID).
		Return(int64(1), nil).Once()

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/links/%d", linkID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, expectedCode, w.Code)
	m.AssertExpectations(t)
}
