package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		count int //передаваемое значение count
		city  string
		want  int // ожидаемое количество кафе в ответе
	}{
		{0, "moscow", 0},
		{1, "moscow", 1},
		{2, "moscow", 2},
		{100, "moscow", len(cafeList["moscow"])},
		{3, "tula", 3},
		{1, "tula", 1},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()

		url := fmt.Sprintf("/cafe?count=%d&city=%s", v.count, v.city)

		req := httptest.NewRequest("Get", url, nil)

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code,
			fmt.Sprintf("For count = %d and city = %s expected status 200, got %d",
				v.count, v.city, response.Code))

		cafeZero := strings.TrimSpace(response.Body.String())
		if cafeZero == "" { // обходим возврат пустой строки
			assert.Equal(t, v.want, 0)
			continue
		}
		cafe := strings.Split(cafeZero, ",")
		assert.Equal(t, v.want, len(cafe),
			fmt.Sprintf("For count = %d and city = %s expected %d cafe, got %d",
				v.count, v.city, v.want, len(cafe)))
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string //передаваемое значение search
		city      string
		wantCount int // ожидаемое кол-во кафе в ответе
	}{
		{"фасоль", "moscow", 0},
		{"кофе", "moscow", 2},
		{"вилка", "moscow", 1},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		url := fmt.Sprintf("/cafe?search=%s&city=%s", v.search, v.city)
		req := httptest.NewRequest("Get", url, nil)

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code,
			fmt.Sprintf("For search = %s and city = %s expected status 200, got %d", v.search, v.city, response.Code))

		body := strings.TrimSpace(response.Body.String())
		cafes := strings.Split(strings.TrimSpace(body), ",")

		//игнорируем пустые строки
		filterCafe := make([]string, 0)
		for _, cafe := range cafes {
			if cafe != "" {
				filterCafe = append(filterCafe, cafe)
			}
		}

		assert.Equal(t, v.wantCount, len(filterCafe),
			fmt.Sprintf("For search = %s and city = %s expected %d cafe, got %d", v.search, v.city, v.wantCount, len(filterCafe)))

		for _, cafe := range filterCafe {
			assert.True(t, strings.Contains(strings.ToLower(cafe), strings.ToLower(v.search)),
				fmt.Sprintf("Cafe '%s' does not contain search string '%s'", cafe, v.search))
		}
	}

}
