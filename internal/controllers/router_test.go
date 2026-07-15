package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterEndpoints(t *testing.T) {
	healthCtrl := NewHealthController()
	downloadCtrl := NewDownloadController(nil, nil)
	router := SetupRouter(healthCtrl, downloadCtrl)

	t.Run("GET /api/health", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/health", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /swagger/index.html redirects or serves 200", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/swagger/index.html", nil)
		req.RequestURI = "/swagger/index.html"
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK && w.Code != http.StatusMovedPermanently {
			t.Errorf("Expected status 200 or 301, got %d", w.Code)
		}
	})
}
