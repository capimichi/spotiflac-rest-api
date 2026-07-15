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
