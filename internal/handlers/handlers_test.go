package handlers_test

import (
	"bytes"
	"code/internal/handlers"
	"code/internal/handlers/mocks"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"code/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setUpRouter(t *testing.T) (*gin.Engine, *mocks.MockLinkService, *mocks.MockVisitService) {
	t.Helper()

	// Создадим моковое хранилище и передадим хендлерам
	linkMock := new(mocks.MockLinkService)
	visitMock := new(mocks.MockVisitService)
	handler := handlers.NewHandler(linkMock, visitMock)

	// Создадим тестовый роутер
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Зададим тестовые маршруты
	router.POST("/api/links", handler.CreateLink)
	router.GET("/api/links", handler.GetLinks)
	router.GET("/api/links/:id", handler.GetLinkByID)
	router.PUT("/api/links/:id", handler.UpdateLinkByID)
	router.DELETE("/api/links/:id", handler.DeleteLinkByID)
	router.GET("/r/:code", handler.RedirectByShortName)
	router.GET("/api/link_visits", handler.GetVisits)

	return router, linkMock, visitMock
}

func TestHandler_CreateLink(t *testing.T) {
	t.Parallel()
	// Arrange
	router, m, _ := setUpRouter(t)

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
	router, m, _ := setUpRouter(t)

	expectedShortUrl1 := "http://localhost:8080/test1"
	expectedShortUrl2 := "http://localhost:8080/test2"

	m.On("GetLinks", mock.Anything, int32(2), int32(0)).Return([]*service.Link{
		{ID: 1, OriginalUrl: "http://test1@gmail.com/long1", ShortName: "test1", ShortUrl: "http://localhost:8080/test1"},
		{ID: 2, OriginalUrl: "http://test2@gmail.com/long2", ShortName: "test2", ShortUrl: "http://localhost:8080/test2"},
	}, int64(2), nil)

	//Act
	req := httptest.NewRequest("GET", "/api/links?range=[0,2]", nil)
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
	router, m, _ := setUpRouter(t)

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
	router, m, _ := setUpRouter(t)

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
	router, m, _ := setUpRouter(t)

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

func TestHandler_RedirectByShortName(t *testing.T) {
	t.Parallel()
	router, linkMock, visitMock := setUpRouter(t)
	shortName := "short"
	expectedOriginalUrl := "https://test1@mail.ru/redirect"
	expectedStatusCode := http.StatusFound

	linkMock.On("GetOriginalURLByShortName", mock.Anything, shortName).
		Return(&service.Link{
			ID:          1,
			OriginalUrl: expectedOriginalUrl,
			ShortName:   shortName,
		}, nil).Once()

	visitMock.On("CreateVisit", mock.Anything, int64(1), "192.0.2.1", "curl/8.14.1", "", int32(302)).
		Return(nil).Once()

	req := httptest.NewRequest("GET", fmt.Sprintf("/r/%s", shortName), nil)
	req.Header.Set("User-Agent", "curl/8.14.1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, expectedStatusCode, w.Code)
	assert.Equal(t, expectedOriginalUrl, w.Header().Get("Location"))
	linkMock.AssertExpectations(t)
	visitMock.AssertExpectations(t)
}

func TestHandler_GetVisits(t *testing.T) {
	t.Parallel()
	router, _, visitMock := setUpRouter(t)

	expected := []*service.Visit{
		{
			Link_ID:   2,
			IP:        "192.168.87.876",
			UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 YaBrowser/24.1.0.0 Safari/537.36",
			Status:    302,
		},
		{
			Link_ID:   2,
			IP:        "192.168.34.189",
			UserAgent: "curl/8.14.1",
			Status:    302,
		},
	}

	visitMock.On("GetVisits", mock.Anything, int32(2), int32(0)).
		Return(expected, int64(2), nil).Once()

	req := httptest.NewRequest("GET", "/api/link_visits?range=[0,2]", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response []*service.Visit
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expected[0], response[0])
	assert.Equal(t, expected[1], response[1])

	visitMock.AssertExpectations(t)
}

func TestSaveConvertToInt32(t *testing.T) {
	t.Parallel()
	var testCases = []struct {
		name  string
		input int
		err   bool
	}{
		{name: "maxInt32_case", input: math.MaxInt32, err: false},
		{name: "minInt32_case", input: math.MinInt32, err: false},
		{name: "greater_than_maxInt32_case", input: math.MaxInt32 + 1, err: true},
		{name: "less_than_minInt32_case", input: math.MinInt32 - 1, err: true},
		{name: "in_range_case", input: 800, err: false},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := handlers.SaveConvertToInt32(tc.input)
			if !tc.err {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestParseAndValidateQuery(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		name       string
		input      string
		wantLimit  int32
		wantOffset int32
		err        bool
	}{
		{name: "successful_input", input: "/api/links?range=[5,10]", wantOffset: 5, wantLimit: 5, err: false},
		{name: "empty_string_input", input: "/api/links", wantOffset: handlers.Default_Offset,
			wantLimit: handlers.Default_Limit, err: false},
		{name: "len_input_more_than_two", input: "/api/links?range=[0,,45]", wantLimit: 0, wantOffset: 0, err: true},
		{name: "range_begin_less_than_zero", input: "/api/links?range=[-4,45]", wantLimit: 0, wantOffset: 0, err: true},
		{name: "range_end_less_than_begin", input: "/api/links?range=[45,4]", wantLimit: 0, wantOffset: 0, err: true},
		{name: "range_end_equal_begin", input: "/api/links?range=[45,45]", wantLimit: 0, wantOffset: 0, err: true},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			url := tc.input
			req := httptest.NewRequest("GET", url, nil)
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			limit, offset, err := handlers.ParseAndValidateQuery(c)
			if !tc.err {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			assert.Equal(t, tc.wantLimit, limit)
			assert.Equal(t, tc.wantOffset, offset)
		})
	}
}

func TestGetIDFromRequest(t *testing.T) {
	t.Parallel()
	id := "15"
	d, err := strconv.Atoi(id)
	require.NoError(t, err)
	want := int64(d)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: id}}
	got := handlers.GetIDFromRequest(c)
	assert.Equal(t, want, got)
}
