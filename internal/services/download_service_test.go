package services

import (
	"testing"
	"time"

	"github.com/capimichi/spotiflac-rest-api/internal/models"
	"github.com/capimichi/spotiflac-rest-api/internal/repositories"
)

func TestDownloadService_DownloadAsync(t *testing.T) {
	repo := repositories.NewInMemoryTaskRepository()
	service := NewDownloadService(repo)

	req := models.DownloadRequest{
		URL:            "https://open.spotify.com/track/4jVnKscdFqT0k2d4Fp4e1F",
		Service:        "qobuz",
		Quality:        "6",
		OutputDir:      "./downloads_test",
		FilenameFormat: "{artist} - {title}",
	}

	task, err := service.DownloadAsync(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if task.ID == "" {
		t.Error("Expected generated task ID")
	}

	// Wait briefly for goroutine starting
	time.Sleep(10 * time.Millisecond)

	retrieved, _, _ := repo.Get(task.ID)
	if retrieved.Status != models.StatusResolving && retrieved.Status != models.StatusFailed {
		t.Errorf("Expected status to be resolving or failed (due to network in test), got %s", retrieved.Status)
	}
}
