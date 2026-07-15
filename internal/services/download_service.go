package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/afkarxyz/SpotiFLAC/backend"
	"github.com/capimichi/spotiflac-rest-api/internal/models"
	"github.com/capimichi/spotiflac-rest-api/internal/repositories"
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
