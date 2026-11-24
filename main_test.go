package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	const city = "tula"
	handler := http.HandlerFunc(mainHandle)
	requests := []struct {
		count int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, min(100, len(cafeList[city]))},
	}

	for _, test := range requests {
		count := strconv.Itoa(test.count)
		t.Run("count="+count, func(t *testing.T) {
			query := url.Values{"city": {city}, "count": {count}}
			req := httptest.NewRequest(http.MethodGet, "/cafe?"+query.Encode(), nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)
			body := strings.TrimSpace(response.Body.String())
			if test.want == 0 {
				assert.Empty(t, body)
				return
			}
			assert.NotEmpty(t, body)
			assert.Len(t, strings.Split(body, ","), test.want)
		})
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
		{"КоФе", 2},
	}

	for _, test := range requests {
		t.Run("search="+test.search, func(t *testing.T) {
			query := url.Values{"city": {"moscow"}, "search": {test.search}}
			req := httptest.NewRequest(http.MethodGet, "/cafe?"+query.Encode(), nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)
			body := strings.TrimSpace(response.Body.String())
			var cafes []string
			if body != "" {
				cafes = strings.Split(body, ",")
			}
			assert.Len(t, cafes, test.wantCount)
			search := strings.ToLower(test.search)
			for _, cafe := range cafes {
				assert.True(t, strings.Contains(strings.ToLower(cafe), search),
					"%q должно содержать %q", cafe, test.search)
			}
		})
	}
}
