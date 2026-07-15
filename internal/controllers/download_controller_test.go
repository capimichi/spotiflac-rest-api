package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/capimichi/spotiflac-rest-api/internal/models"
	"github.com/capimichi/spotiflac-rest-api/internal/repositories"
	"github.com/capimichi/spotiflac-rest-api/internal/services"
)

func TestDownloadControllerEndpoints(t *testing.T) {
	repo := repositories.NewInMemoryTaskRepository()
	svc := services.NewDownloadService(repo)
	healthCtrl := NewHealthController()
	downloadCtrl := NewDownloadController(svc, repo)
	router := SetupRouter(healthCtrl, downloadCtrl)

	t.Run("POST /api/download bad request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/download", bytes.NewBufferString("{invalid-json}"))
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("GET /api/status/:id not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/status/non-existent-id", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("POST /api/download valid async request structure", func(t *testing.T) {
		w := httptest.NewRecorder()
		reqData := models.DownloadRequest{
			URL: "https://open.spotify.com/track/4jVnKscdFqT0k2d4Fp4e1F",
		}
		jsonVal, _ := json.Marshal(reqData)
		req, _ := http.NewRequest("POST", "/api/download", bytes.NewBuffer(jsonVal))
		router.ServeHTTP(w, req)

		if w.Code != http.StatusAccepted && w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 202 or 500, got %d", w.Code)
		}
	})
}
