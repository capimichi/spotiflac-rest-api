package models

type DownloadRequest struct {
	// URL del brano, album, playlist o artista da scaricare. Supporta link Spotify (es. https://open.spotify.com/track/...) oppure link diretti Tidal, Amazon Music o Qobuz. Obbligatorio a meno che non si specifichi la triade isrc, track_name e artist_name.
	URL            string `json:"url" example:"https://open.spotify.com/track/4PTG3Z6ehGkBFmzsOhswXP"`

	// Servizio di streaming musicale di destinazione da cui effettuare il download effettivo dei file.
	// @Enums qobuz, tidal, amazon
	Service        string `json:"service" example:"qobuz"`

	// Qualità del download. Per Qobuz (es. "6" per Hi-Res). Per Tidal/Amazon (es. "LOSSLESS" per qualità CD FLAC). Se omesso, viene applicato un valore predefinito ottimale.
	Quality        string `json:"quality" example:"6"`

	// Cartella sul server in cui salvare i file musicali scaricati. Default: "./downloads"
	OutputDir      string `json:"output_dir" example:"./downloads"`

	// Formato di salvataggio per il nome del file. Supporta tag dinamici come {artist}, {title}, {album}. Default: "{artist} - {title}"
	FilenameFormat string `json:"filename_format" example:"{artist} - {title}"`

	// Codice ISRC (International Standard Recording Code) del brano. Obbligatorio solo per il download diretto tramite metadati (se non viene fornito un URL).
	ISRC           string `json:"isrc" example:"USUM71703861"`

	// Titolo del brano. Obbligatorio solo se non viene fornito un URL. Nei link diretti, serve come metadato di fallback.
	TrackName      string `json:"track_name" example:"Look What You Made Me Do"`

	// Nome dell'artista principale. Obbligatorio solo se non viene fornito un URL. Nei link diretti, serve come metadato di fallback.
	ArtistName     string `json:"artist_name" example:"Taylor Swift"`

	// Nome dell'album. Facoltativo, usato per la scrittura dei tag ID3/FLAC nei download da link diretti o ricerca ISRC.
	AlbumName      string `json:"album_name" example:"Reputation"`

	// Artista dell'album. Facoltativo, usato per la scrittura dei tag ID3/FLAC nei download da link diretti o ricerca ISRC.
	AlbumArtist    string `json:"album_artist" example:"Taylor Swift"`

	// Data di pubblicazione del brano (es. YYYY-MM-DD). Facoltativo, usato per i tag ID3/FLAC.
	ReleaseDate    string `json:"release_date" example:"2017-08-24"`

	// ID Spotify del brano. Facoltativo, utile come riferimento per collegare l'ISRC in caso di download diretto senza URL.
	SpotifyID      string `json:"spotify_id" example:"1OCaN2Vv7f7F1QGq5Q7P9H"`
}

