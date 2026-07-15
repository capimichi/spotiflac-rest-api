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
