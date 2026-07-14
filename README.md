# SpotiFLAC REST API

A lightweight REST API wrapper for the [SpotiFLAC](https://github.com/afkarxyz/SpotiFLAC) backend. This service allows you to download Spotify tracks, albums, and playlists as high-quality lossless FLAC (or MP3/AAC) files via Qobuz, Tidal, or Amazon Music.

## Prerequisites

* Go 1.26 or higher
* `ffmpeg` installed on your system (required by the SpotiFLAC backend for post-processing and tagging)

## Getting Started

1. **Run the server**:
   ```bash
   go run main.go
   ```
   By default, the server runs on port `8080`.

2. **Change port (optional)**:
   ```bash
   PORT=9000 go run main.go
   ```

---

## API Documentation

### 1. Health Check
Checks if the API server is online and running.
* **Method**: `GET`
* **Path**: `/api/health`
* **Response**:
  ```json
  {
    "status": "ok",
    "time": "2026-07-14T23:15:00+02:00",
    "environment": "debug"
  }
  ```

### 2. Async Download
Starts a download task in the background and returns a task ID immediately. This prevents HTTP timeouts on large playlists.
* **Method**: `POST`
* **Path**: `/api/download`
* **Headers**: `Content-Type: application/json`
* **Payload**:
  ```json
  {
    "url": "https://open.spotify.com/track/4jVnKscdFqT0k2d4Fp4e1F",
    "service": "qobuz",
    "quality": "6",
    "output_dir": "./downloads",
    "filename_format": "{artist} - {title}"
  }
  ```
  * `service` options: `"qobuz"` (default), `"tidal"`, `"amazon"`
  * `quality` options: 
    * Qobuz: `"6"` (16-bit Lossless, default), `"7"` / `"27"` (Hi-Res)
    * Tidal: `"LOSSLESS"` (default), `"HI_RES"`
* **Response**:
  ```json
  {
    "task_id": "c1f8d9b2a7c4f6b2",
    "status": "pending",
    "message": "Download task started successfully"
  }
  ```

### 3. Track Task Status
Query the status of an asynchronous download task.
* **Method**: `GET`
* **Path**: `/api/status/:id`
* **Response (Downloading)**:
  ```json
  {
    "id": "c1f8d9b2a7c4f6b2",
    "spotify_url": "https://open.spotify.com/track/4jVnKscdFqT0k2d4Fp4e1F",
    "service": "qobuz",
    "quality": "6",
    "status": "downloading",
    "current_track": "Macklemore - 1984",
    "completed_tracks": 0,
    "total_tracks": 1
  }
  ```
* **Response (Completed)**:
  ```json
  {
    "id": "c1f8d9b2a7c4f6b2",
    "spotify_url": "https://open.spotify.com/track/4jVnKscdFqT0k2d4Fp4e1F",
    "service": "qobuz",
    "quality": "6",
    "status": "completed",
    "completed_tracks": 1,
    "total_tracks": 1,
    "downloaded_files": [
      "downloads/Macklemore - 1984.flac"
    ]
  }
  ```

### 4. Synchronous Download
Blocks until the entire download process (including tagging) is complete, then returns the saved files. Useful for scripting.
* **Method**: `POST`
* **Path**: `/api/download/sync`
* **Payload**: Same as `/api/download`
* **Response**:
  ```json
  {
    "status": "completed",
    "files": [
      "downloads/Macklemore - 1984.flac"
    ]
  }
  ```
