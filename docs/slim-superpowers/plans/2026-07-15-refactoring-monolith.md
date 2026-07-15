# Refactoring Monolith to MVC/Multi-Layer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactoring the monolithic `main.go` file into a multi-layer architectural pattern (Models, Repositories, Services, Controllers) using Standard Go Project Layout (`cmd/` and `internal/`) to improve maintainability, testability, and separation of concerns.

**Architecture:** We will decompose the application into four distinct packages inside `internal/`: `models` (data types), `repositories` (persistence contract and RAM implementation), `services` (business logic for Spotify metadata and SpotiFLAC downloader), and `controllers` (http endpoints and routing setup). The entrypoint will be moved to `cmd/server/main.go`.

**Tech Stack:** Go (Golang) >= 1.26, Gin Web Framework, SpotiFLAC backend package.

## Global Constraints
- Go Version: 1.26 or higher
- Web Framework: Gin
- Dependency: SpotiFLAC backend (`github.com/afkarxyz/SpotiFLAC/backend`)
- Target Structure: `cmd/server/main.go` for entrypoint, `internal/` directory for core packages.

---

### Task 1: Scaffolding and Models Creation

**Files:**
- Create: `internal/models/task.go`
- Create: `internal/models/request.go`
- Test: We will verify compiler verification and packages separation.

**Interfaces:**
- Produces: `models.TaskStatus`, `models.Task` struct, `models.DownloadRequest` struct.

- [ ] **Step 1: Create the task status and structure file**
Create the file `internal/models/task.go` containing:
```go
package models

import "time"

type TaskStatus string

const (
	StatusPending     TaskStatus = "pending"
	StatusResolving   TaskStatus = "resolving"
	StatusDownloading TaskStatus = "downloading"
	StatusCompleted   TaskStatus = "completed"
	StatusFailed      TaskStatus = "failed"
)

type Task struct {
	ID              string     `json:"id"`
	SpotifyURL      string     `json:"spotify_url"`
	Service         string     `json:"service"`
	Quality         string     `json:"quality"`
	Status          TaskStatus `json:"status"`
	CurrentTrack    string     `json:"current_track,omitempty"`
	CompletedTracks int        `json:"completed_tracks"`
	TotalTracks     int        `json:"total_tracks"`
	DownloadedFiles []string   `json:"downloaded_files,omitempty"`
	Error           string     `json:"error,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
```

- [ ] **Step 2: Create the request model file**
Create the file `internal/models/request.go` containing:
```go
package models

type DownloadRequest struct {
	URL            string `json:"url"`
	Service        string `json:"service"`
	Quality        string `json:"quality"`
	OutputDir      string `json:"output_dir"`
	FilenameFormat string `json:"filename_format"`
	ISRC           string `json:"isrc"`
	TrackName      string `json:"track_name"`
	ArtistName     string `json:"artist_name"`
	AlbumName      string `json:"album_name"`
	AlbumArtist    string `json:"album_artist"`
	ReleaseDate    string `json:"release_date"`
	SpotifyID      string `json:"spotify_id"`
}
```

- [ ] **Step 3: Verify the code compiles**
Run: `go build ./internal/models`
Expected: Command exits successfully with code 0 (no compile errors).

---

### Task 2: Create the Task Repository layer

**Files:**
- Create: `internal/repositories/repository.go`
- Create: `internal/repositories/memory.go`
- Test: `internal/repositories/memory_test.go`

**Interfaces:**
- Consumes: `models.Task`, `models.TaskStatus`
- Produces: `repositories.TaskRepository` interface, `repositories.NewInMemoryTaskRepository()` function.

- [ ] **Step 1: Write the failing unit tests for memory repository**
Create `internal/repositories/memory_test.go` containing:
```go
package repositories

import (
	"testing"
	"spotiflac-rest-api/internal/models"
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
```

- [ ] **Step 2: Run tests to verify they fail to compile**
Run: `go test -v ./internal/repositories`
Expected: Compile failure because `NewInMemoryTaskRepository` is undefined.

- [ ] **Step 3: Implement TaskRepository Interface**
Create `internal/repositories/repository.go` containing:
```go
package repositories

import "spotiflac-rest-api/internal/models"

type TaskRepository interface {
	Create(spotifyURL, service, quality string) (*models.Task, error)
	Get(id string) (*models.Task, bool, error)
	Update(id string, fn func(*models.Task)) error
}
```

- [ ] **Step 4: Implement InMemoryTaskRepository**
Create `internal/repositories/memory.go` containing:
```go
package repositories

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"spotiflac-rest-api/internal/models"
)

type InMemoryTaskRepository struct {
	mu    sync.RWMutex
	tasks map[string]*models.Task
}

func NewInMemoryTaskRepository() *InMemoryTaskRepository {
	return &InMemoryTaskRepository{
		tasks: make(map[string]*models.Task),
	}
}

func (r *InMemoryTaskRepository) Create(spotifyURL, service, quality string) (*models.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := generateID()
	task := &models.Task{
		ID:         id,
		SpotifyURL: spotifyURL,
		Service:    service,
		Quality:    quality,
		Status:     models.StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	r.tasks[id] = task
	return task, nil
}

func (r *InMemoryTaskRepository) Get(id string) (*models.Task, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.tasks[id]
	if !exists {
		return nil, false, nil
	}
	taskCopy := *task
	return &taskCopy, true, nil
}

func (r *InMemoryTaskRepository) Update(id string, fn func(*models.Task)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if task, exists := r.tasks[id]; exists {
		fn(task)
		task.UpdatedAt = time.Now()
	}
	return nil
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
```

- [ ] **Step 5: Run tests to verify they pass**
Run: `go test -v ./internal/repositories`
Expected: PASS

---

### Task 3: Create Download Service

**Files:**
- Create: `internal/services/download_service.go`
- Create: `internal/services/download_service_test.go`

**Interfaces:**
- Consumes: `repositories.TaskRepository`
- Produces: `services.NewDownloadService(repo repositories.TaskRepository)` function, and methods `DownloadAsync` and `DownloadSync`.

- [ ] **Step 1: Write unit tests for DownloadService**
Create `internal/services/download_service_test.go` containing:
```go
package services

import (
	"testing"
	"time"

	"spotiflac-rest-api/internal/models"
	"spotiflac-rest-api/internal/repositories"
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
```

- [ ] **Step 2: Run test to verify compile failure**
Run: `go test -v ./internal/services`
Expected: Compile failure due to missing `NewDownloadService`.

- [ ] **Step 3: Implement DownloadService**
Create `internal/services/download_service.go` containing:
```go
package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/afkarxyz/SpotiFLAC/backend"
	"spotiflac-rest-api/internal/models"
	"spotiflac-rest-api/internal/repositories"
)

type DownloadService struct {
	repo repositories.TaskRepository
}

func NewDownloadService(repo repositories.TaskRepository) *DownloadService {
	return &DownloadService{
		repo: repo,
	}
}

func (s *DownloadService) DownloadAsync(req models.DownloadRequest) (*models.Task, error) {
	task, err := s.repo.Create(req.URL, req.Service, req.Quality)
	if err != nil {
		return nil, err
	}

	go s.downloadTask(task.ID, req)

	return task, nil
}

func (s *DownloadService) DownloadSync(req models.DownloadRequest) ([]string, error) {
	return s.executeDownload(req, nil)
}

func (s *DownloadService) downloadTask(taskID string, req models.DownloadRequest) {
	_ = s.repo.Update(taskID, func(t *models.Task) {
		t.Status = models.StatusResolving
	})

	files, err := s.executeDownload(req, func(currentTrack string, completed, total int) {
		_ = s.repo.Update(taskID, func(t *models.Task) {
			t.Status = models.StatusDownloading
			t.CurrentTrack = currentTrack
			t.CompletedTracks = completed
			t.TotalTracks = total
		})
	})

	if err != nil {
		_ = s.repo.Update(taskID, func(t *models.Task) {
			t.Status = models.StatusFailed
			t.Error = err.Error()
		})
		return
	}

	_ = s.repo.Update(taskID, func(t *models.Task) {
		t.Status = models.StatusCompleted
		t.DownloadedFiles = files
		t.CurrentTrack = ""
	})
}

type trackMetadata struct {
	SpotifyID          string
	TrackName          string
	ArtistName         string
	AlbumName          string
	AlbumArtist        string
	ReleaseDate        string
	CoverURL           string
	TrackNumber        int
	DiscNumber         int
	TotalTracks        int
	TotalDiscs         int
	UPC                string
	Copyright          string
	Publisher          string
	Composer           string
	SpotifyURL         string
	DurationMS         int
}

type progressCallback func(currentTrack string, completed, total int)

func (s *DownloadService) executeDownload(req models.DownloadRequest, progress progressCallback) ([]string, error) {
	isTidal := strings.Contains(req.URL, "tidal.")
	isAmazon := strings.Contains(req.URL, "amazon.")
	isQobuz := strings.Contains(req.URL, "qobuz.")

	if isAmazon && strings.Contains(req.URL, "trackAsin=") {
		parts := strings.Split(req.URL, "trackAsin=")
		if len(parts) > 1 {
			extractedASIN := ""
			for _, r := range parts[1] {
				if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') {
					extractedASIN += string(r)
				} else {
					break
				}
			}
			if len(extractedASIN) == 10 {
				req.URL = "https://music.amazon.com/tracks/" + extractedASIN
			}
		}
	}

	if isTidal || isAmazon || isQobuz {
		var filename string
		var dlErr error

		trackName := req.TrackName
		if trackName == "" {
			trackName = "Unknown Track"
		}
		artistName := req.ArtistName
		if artistName == "" {
			artistName = "Unknown Artist"
		}

		if isQobuz {
			qobuzID := extractQobuzTrackID(req.URL)
			if qobuzID == "" {
				return nil, fmt.Errorf("could not extract Qobuz track ID from URL: %s", req.URL)
			}
			isrcVal := "qobuz_" + qobuzID
			
			downloader := backend.NewQobuzDownloader()
			filename, dlErr = downloader.DownloadTrackWithISRC(
				isrcVal,
				req.OutputDir,
				req.Quality,
				req.FilenameFormat,
				false,
				1,
				trackName,
				artistName,
				req.AlbumName,
				req.AlbumArtist,
				req.ReleaseDate,
				false,
				"",
				true,
				1,
				1,
				1,
				1,
				"",
				"",
				"",
				", ",
				"",
				true,
				false,
				false,
				false,
			)
		} else if isTidal {
			downloader := backend.NewTidalDownloader("")
			filename, dlErr = downloader.DownloadByURL(
				req.URL,
				req.OutputDir,
				req.Quality,
				req.FilenameFormat,
				false,
				1,
				trackName,
				artistName,
				req.AlbumName,
				req.AlbumArtist,
				req.ReleaseDate,
				false,
				"",
				true,
				1,
				1,
				1,
				1,
				"",
				"",
				"",
				", ",
				req.ISRC,
				"",
				true,
				false,
				false,
				false,
			)
		} else if isAmazon {
			downloader := backend.NewAmazonDownloader()
			filename, dlErr = downloader.DownloadByURL(
				req.URL,
				req.OutputDir,
				req.Quality,
				req.FilenameFormat,
				"",
				"",
				false,
				1,
				trackName,
				artistName,
				req.AlbumName,
				req.AlbumArtist,
				req.ReleaseDate,
				"",
				1,
				1,
				1,
				true,
				1,
				"",
				"",
				"",
				", ",
				req.ISRC,
				"",
				false,
				false,
				false,
			)
		}

		if dlErr != nil {
			return nil, dlErr
		}
		return []string{filename}, nil
	}

	var tracks []trackMetadata

	if req.ISRC != "" && req.TrackName != "" && req.ArtistName != "" {
		spotifyID := req.SpotifyID
		if spotifyID == "" && req.URL != "" {
			parts := strings.Split(req.URL, "/")
			if len(parts) > 0 {
				lastPart := parts[len(parts)-1]
				spotifyID = strings.Split(lastPart, "?")[0]
			}
		}

		tracks = append(tracks, trackMetadata{
			SpotifyID:   spotifyID,
			TrackName:   req.TrackName,
			ArtistName:  req.ArtistName,
			AlbumName:   req.AlbumName,
			AlbumArtist: req.AlbumArtist,
			ReleaseDate: req.ReleaseDate,
			TrackNumber: 1,
			DiscNumber:  1,
			TotalTracks: 1,
			TotalDiscs:  1,
			UPC:         req.ISRC,
			SpotifyURL:  req.URL,
		})
	} else {
		if req.URL == "" {
			return nil, fmt.Errorf("either 'url' or direct metadata ('isrc', 'track_name', 'artist_name') must be provided")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		data, err := backend.GetFilteredSpotifyData(ctx, req.URL, false, 0, ", ", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch Spotify metadata: %w", err)
		}

		switch res := data.(type) {
		case backend.TrackResponse:
			tracks = append(tracks, mapTrackMetadata(res.Track))

		case *backend.AlbumResponsePayload:
			for _, item := range res.TrackList {
				tracks = append(tracks, mapAlbumTrackMetadata(item, res.AlbumInfo.UPC))
			}

		case *backend.PlaylistResponsePayload:
			for _, item := range res.TrackList {
				tracks = append(tracks, mapAlbumTrackMetadata(item, ""))
			}

		case *backend.ArtistDiscographyPayload:
			for _, item := range res.TrackList {
				tracks = append(tracks, mapAlbumTrackMetadata(item, ""))
			}

		default:
			return nil, fmt.Errorf("unsupported Spotify metadata response type: %T", data)
		}
	}

	total := len(tracks)
	if total == 0 {
		return nil, fmt.Errorf("no tracks found to download")
	}

	var downloadedFiles []string

	for idx, track := range tracks {
		trackDisplay := fmt.Sprintf("%s - %s", track.ArtistName, track.TrackName)
		if progress != nil {
			progress(trackDisplay, idx, total)
		}

		isrc := backend.ResolveTrackISRC(track.SpotifyID)
		if isrc == "" {
			isrc = track.UPC
		}

		var filename string
		var dlErr error

		switch req.Service {
		case "qobuz":
			downloader := backend.NewQobuzDownloader()
			filename, dlErr = downloader.DownloadTrackWithISRC(
				isrc,
				req.OutputDir,
				req.Quality,
				req.FilenameFormat,
				false,
				idx+1,
				track.TrackName,
				track.ArtistName,
				track.AlbumName,
				track.AlbumArtist,
				track.ReleaseDate,
				false,
				track.CoverURL,
				true,
				track.TrackNumber,
				track.DiscNumber,
				track.TotalTracks,
				track.TotalDiscs,
				track.Copyright,
				track.Publisher,
				track.Composer,
				", ",
				track.SpotifyURL,
				true,
				false,
				false,
				false,
			)

		case "tidal":
			downloader := backend.NewTidalDownloader("")
			filename, dlErr = downloader.Download(
				track.SpotifyID,
				req.OutputDir,
				req.Quality,
				req.FilenameFormat,
				false,
				idx+1,
				track.TrackName,
				track.ArtistName,
				track.AlbumName,
				track.AlbumArtist,
				track.ReleaseDate,
				false,
				track.CoverURL,
				true,
				track.TrackNumber,
				track.DiscNumber,
				track.TotalTracks,
				track.TotalDiscs,
				track.Copyright,
				track.Publisher,
				track.Composer,
				", ",
				isrc,
				track.SpotifyURL,
				true,
				false,
				false,
				false,
			)

		case "amazon":
			downloader := backend.NewAmazonDownloader()
			filename, dlErr = downloader.DownloadBySpotifyID(
				track.SpotifyID,
				req.OutputDir,
				req.Quality,
				req.FilenameFormat,
				"",
				"",
				false,
				idx+1,
				track.TrackName,
				track.ArtistName,
				track.AlbumName,
				track.AlbumArtist,
				track.ReleaseDate,
				track.CoverURL,
				track.TrackNumber,
				track.DiscNumber,
				track.TotalTracks,
				true,
				track.TotalDiscs,
				track.Copyright,
				track.Publisher,
				track.Composer,
				", ",
				isrc,
				track.SpotifyURL,
				false,
				false,
				false,
			)

		default:
			return downloadedFiles, fmt.Errorf("unknown service: %s", req.Service)
		}

		if dlErr != nil {
			continue
		}

		downloadedFiles = append(downloadedFiles, filename)
	}

	if progress != nil {
		progress("", total, total)
	}

	return downloadedFiles, nil
}

func mapTrackMetadata(t backend.TrackMetadata) trackMetadata {
	return trackMetadata{
		SpotifyID:   t.SpotifyID,
		TrackName:   t.Name,
		ArtistName:  t.Artists,
		AlbumName:   t.AlbumName,
		AlbumArtist: t.AlbumArtist,
		ReleaseDate: t.ReleaseDate,
		CoverURL:    t.Images,
		TrackNumber: t.TrackNumber,
		DiscNumber:  t.DiscNumber,
		TotalTracks: t.TotalTracks,
		TotalDiscs:  t.TotalDiscs,
		UPC:         t.UPC,
		Copyright:   t.Copyright,
		Publisher:   t.Publisher,
		Composer:    t.Composer,
		SpotifyURL:  t.ExternalURL,
		DurationMS:  t.DurationMS,
	}
}

func mapAlbumTrackMetadata(t backend.AlbumTrackMetadata, upc string) trackMetadata {
	actualUPC := t.UPC
	if actualUPC == "" {
		actualUPC = upc
	}
	return trackMetadata{
		SpotifyID:   t.SpotifyID,
		TrackName:   t.Name,
		ArtistName:  t.Artists,
		AlbumName:   t.AlbumName,
		AlbumArtist: t.AlbumArtist,
		ReleaseDate: t.ReleaseDate,
		CoverURL:    t.Images,
		TrackNumber: t.TrackNumber,
		DiscNumber:  t.DiscNumber,
		TotalTracks: t.TotalTracks,
		TotalDiscs:  t.TotalDiscs,
		UPC:         actualUPC,
		SpotifyURL:  t.ExternalURL,
		DurationMS:  t.DurationMS,
	}
}

func extractQobuzTrackID(qobuzURL string) string {
	if strings.Contains(qobuzURL, "track_id=") {
		parts := strings.Split(qobuzURL, "track_id=")
		if len(parts) > 1 {
			id := ""
			for _, r := range parts[1] {
				if r >= '0' && r <= '9' {
					id += string(r)
				} else {
					break
				}
			}
			if id != "" {
				return id
			}
		}
	}
	if strings.Contains(qobuzURL, "/track/") {
		parts := strings.Split(qobuzURL, "/track/")
		if len(parts) > 1 {
			id := ""
			for _, r := range parts[1] {
				if r >= '0' && r <= '9' {
					id += string(r)
				} else {
					break
				}
			}
			if id != "" {
				return id
			}
		}
	}
	return ""
}
