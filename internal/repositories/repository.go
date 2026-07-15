package repositories

import "github.com/capimichi/spotiflac-rest-api/internal/models"

type TaskRepository interface {
	Create(spotifyURL, service, quality string) (*models.Task, error)
	Get(id string) (*models.Task, bool, error)
	Update(id string, fn func(*models.Task)) error
}
