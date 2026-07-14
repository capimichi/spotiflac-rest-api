package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/afkarxyz/SpotiFLAC/backend"
	"github.com/gin-gonic/gin"
)

// TaskStatus represents the status of a download task.
type TaskStatus string

const (
	StatusPending     TaskStatus = "pending"
	StatusResolving   TaskStatus = "resolving"
	StatusDownloading TaskStatus = "downloading"
	StatusCompleted   TaskStatus = "completed"
	StatusFailed      TaskStatus = "failed"
)

// Task represents a background download task.
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

// TaskStore manages tasks in memory.
type TaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks: make(map[string]*Task),
	}
}

func (s *TaskStore) Create(spotifyURL, service, quality string) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateID()
	task := &Task{
		ID:         id,
		SpotifyURL: spotifyURL,
		Service:    service,
		Quality:    quality,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	s.tasks[id] = task
	return task
}

func (s *TaskStore) Get(id string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, false
	}
	// Return a copy to avoid race conditions when reading
	taskCopy := *task
	return &taskCopy, true
}

func (s *TaskStore) Update(id string, fn func(*Task)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.tasks[id]; exists {
		fn(task)
		task.UpdatedAt = time.Now()
	}
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Global store
var store = NewTaskStore()

type DownloadRequest struct {
	URL            string `json:"url" binding:"required"`
	Service        string `json:"service"` // "qobuz", "tidal", "amazon"
	Quality        string `json:"quality"` // Qobuz: "6" (Lossless), "7"/"27" (Hi-Res) | Tidal: "HI_RES", "LOSSLESS"
	OutputDir      string `json:"output_dir"`
	FilenameFormat string `json:"filename_format"`
}

func main() {
	// Enable release mode for Gin if requested, default to debug for development
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health Check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"time":        time.Now().Format(time.RFC3339),
			"environment": gin.Mode(),
		})
	})

	// Get Task Status
	r.GET("/api/status/:id", func(c *gin.Context) {
		id := c.Param("id")
		task, exists := store.Get(id)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusOK, task)
	})

	// Async Download Endpoint
	r.POST("/api/download", func(c *gin.Context) {
		var req DownloadRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Defaults
		if req.Service == "" {
			req.Service = "qobuz"
		}
		if req.Quality == "" {
			if req.Service == "qobuz" {
				req.Quality = "6" // 16-bit Lossless
			} else if req.Service == "tidal" {
				req.Quality = "LOSSLESS"
			} else {
				req.Quality = "LOSSLESS"
			}
		}
		if req.OutputDir == "" {
			req.OutputDir = "./downloads"
		}
		if req.FilenameFormat == "" {
			req.FilenameFormat = "{artist} - {title}"
		}

		task := store.Create(req.URL, req.Service, req.Quality)

		// Trigger download in background
		go downloadTask(task.ID, req)

		c.JSON(http.StatusAccepted, gin.H{
			"task_id": task.ID,
			"status":  task.Status,
			"message": "Download task started successfully",
		})
	})

	// Sync Download Endpoint (blocks until done)
	r.POST("/api/download/sync", func(c *gin.Context) {
		var req DownloadRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Defaults
		if req.Service == "" {
			req.Service = "qobuz"
		}
		if req.Quality == "" {
			if req.Service == "qobuz" {
				req.Quality = "6"
			} else {
				req.Quality = "LOSSLESS"
			}
		}
		if req.OutputDir == "" {
			req.OutputDir = "./downloads"
		}
		if req.FilenameFormat == "" {
			req.FilenameFormat = "{artist} - {title}"
		}

		files, err := executeDownload(req, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "failed",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "completed",
			"files":  files,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("SpotiFLAC REST API Server listening on port %s...\n", port)
	if err := r.Run(":" + port); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
	}
}

// downloadTask manages the lifecycle of the task in background
func downloadTask(taskID string, req DownloadRequest) {
	store.Update(taskID, func(t *Task) {
		t.Status = StatusResolving
	})

	files, err := executeDownload(req, func(currentTrack string, completed, total int) {
		store.Update(taskID, func(t *Task) {
			t.Status = StatusDownloading
			t.CurrentTrack = currentTrack
			t.CompletedTracks = completed
			t.TotalTracks = total
		})
	})

	if err != nil {
		store.Update(taskID, func(t *Task) {
			t.Status = StatusFailed
			t.Error = err.Error()
		})
		return
	}

	store.Update(taskID, func(t *Task) {
		t.Status = StatusCompleted
		t.DownloadedFiles = files
		t.CurrentTrack = ""
	})
}

// Single item representing unified track metadata
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

func executeDownload(req DownloadRequest, progress progressCallback) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	fmt.Printf("Fetching Spotify metadata for URL: %s\n", req.URL)
	data, err := backend.GetFilteredSpotifyData(ctx, req.URL, false, 0, ", ", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Spotify metadata: %w", err)
	}

	var tracks []trackMetadata

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

	total := len(tracks)
	if total == 0 {
		return nil, fmt.Errorf("no tracks found to download")
	}

	fmt.Printf("Resolved %d tracks to download\n", total)
	var downloadedFiles []string

	for idx, track := range tracks {
		trackDisplay := fmt.Sprintf("%s - %s", track.ArtistName, track.TrackName)
		if progress != nil {
			progress(trackDisplay, idx, total)
		}

		fmt.Printf("[%d/%d] Downloading: %s\n", idx+1, total, trackDisplay)
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
				false, // includeTrackNumber (usually handled in filename format)
				idx+1, // position
				track.TrackName,
				track.ArtistName,
				track.AlbumName,
				track.AlbumArtist,
				track.ReleaseDate,
				false, // useAlbumTrackNumber
				track.CoverURL,
				true,  // embedMaxQualityCover
				track.TrackNumber,
				track.DiscNumber,
				track.TotalTracks,
				track.TotalDiscs,
				track.Copyright,
				track.Publisher,
				track.Composer,
				", ", // metadataSeparator
				track.SpotifyURL,
				true,  // allowFallback
				false, // useFirstArtistOnly
				false, // useSingleGenre
				false, // embedGenre
			)

		case "tidal":
			downloader := backend.NewTidalDownloader("")
			filename, dlErr = downloader.Download(
				track.SpotifyID,
				req.OutputDir,
				req.Quality,
				req.FilenameFormat,
				false, // includeTrackNumber
				idx+1, // position
				track.TrackName,
				track.ArtistName,
				track.AlbumName,
				track.AlbumArtist,
				track.ReleaseDate,
				false, // useAlbumTrackNumber
				track.CoverURL,
				true,  // embedMaxQualityCover
				track.TrackNumber,
				track.DiscNumber,
				track.TotalTracks,
				track.TotalDiscs,
				track.Copyright,
				track.Publisher,
				track.Composer,
				", ", // metadataSeparator
				isrc,
				track.SpotifyURL,
				true,  // allowFallback
				false, // useFirstArtistOnly
				false, // useSingleGenre
				false, // embedGenre
			)

		case "amazon":
			downloader := backend.NewAmazonDownloader()
			filename, dlErr = downloader.DownloadBySpotifyID(
				track.SpotifyID,
				req.OutputDir,
				req.Quality,
				req.FilenameFormat,
				"", // playlistName
				"", // playlistOwner
				false, // includeTrackNumber
				idx+1, // position
				track.TrackName,
				track.ArtistName,
				track.AlbumName,
				track.AlbumArtist,
				track.ReleaseDate,
				track.CoverURL,
				track.TrackNumber,
				track.DiscNumber,
				track.TotalTracks,
				true,  // embedMaxQualityCover
				track.TotalDiscs,
				track.Copyright,
				track.Publisher,
				track.Composer,
				", ", // metadataSeparator
				isrc,
				track.SpotifyURL,
				false, // useFirstArtistOnly
				false, // useSingleGenre
				false, // embedGenre
			)

		default:
			return downloadedFiles, fmt.Errorf("unknown service: %s", req.Service)
		}

		if dlErr != nil {
			fmt.Printf("Warning: failed to download track %s: %v\n", trackDisplay, dlErr)
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
		SpotifyID:          t.SpotifyID,
		TrackName:          t.Name,
		ArtistName:         t.Artists,
		AlbumName:          t.AlbumName,
		AlbumArtist:        t.AlbumArtist,
		ReleaseDate:        t.ReleaseDate,
		CoverURL:           t.Images,
		TrackNumber:        t.TrackNumber,
		DiscNumber:         t.DiscNumber,
		TotalTracks:        t.TotalTracks,
		TotalDiscs:         t.TotalDiscs,
		UPC:                t.UPC,
		Copyright:          t.Copyright,
		Publisher:          t.Publisher,
		Composer:           t.Composer,
		SpotifyURL:         t.ExternalURL,
		DurationMS:         t.DurationMS,
	}
}

func mapAlbumTrackMetadata(t backend.AlbumTrackMetadata, upc string) trackMetadata {
	actualUPC := t.UPC
	if actualUPC == "" {
		actualUPC = upc
	}
	return trackMetadata{
		SpotifyID:          t.SpotifyID,
		TrackName:          t.Name,
		ArtistName:         t.Artists,
		AlbumName:          t.AlbumName,
		AlbumArtist:        t.AlbumArtist,
		ReleaseDate:        t.ReleaseDate,
		CoverURL:           t.Images,
		TrackNumber:        t.TrackNumber,
		DiscNumber:         t.DiscNumber,
		TotalTracks:        t.TotalTracks,
		TotalDiscs:         t.TotalDiscs,
		UPC:                actualUPC,
		SpotifyURL:         t.ExternalURL,
		DurationMS:         t.DurationMS,
	}
}
