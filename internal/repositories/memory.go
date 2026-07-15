package repositories

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/capimichi/spotiflac-rest-api/internal/models"
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
