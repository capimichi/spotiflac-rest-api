package repositories

import (
	"testing"

	"github.com/capimichi/spotiflac-rest-api/internal/models"
)

func TestInMemoryTaskRepository_CreateAndGet(t *testing.T) {
	repo := NewInMemoryTaskRepository()

	task, err := repo.Create("https://open.spotify.com/track/123", "qobuz", "6")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if task.ID == "" {
		t.Error("Expected task ID to be generated")
	}

	retrieved, exists, err := repo.Get(task.ID)
	if err != nil {
		t.Fatalf("Expected no error on Get, got %v", err)
	}
	if !exists {
		t.Fatalf("Expected task to exist")
	}

	if retrieved.SpotifyURL != "https://open.spotify.com/track/123" {
		t.Errorf("Expected URL https://open.spotify.com/track/123, got %s", retrieved.SpotifyURL)
	}
}

func TestInMemoryTaskRepository_Update(t *testing.T) {
	repo := NewInMemoryTaskRepository()

	task, _ := repo.Create("https://open.spotify.com/track/123", "qobuz", "6")

	err := repo.Update(task.ID, func(tk *models.Task) {
		tk.Status = models.StatusDownloading
		tk.CompletedTracks = 2
	})
	if err != nil {
		t.Fatalf("Expected no error on Update, got %v", err)
	}

	retrieved, _, _ := repo.Get(task.ID)
	if retrieved.Status != models.StatusDownloading {
		t.Errorf("Expected status %s, got %s", models.StatusDownloading, retrieved.Status)
	}
	if retrieved.CompletedTracks != 2 {
		t.Errorf("Expected CompletedTracks to be 2, got %d", retrieved.CompletedTracks)
	}
}
