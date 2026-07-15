package controllers

import (
	"net/http"

	"github.com/capimichi/spotiflac-rest-api/internal/models"
	"github.com/capimichi/spotiflac-rest-api/internal/repositories"
	"github.com/capimichi/spotiflac-rest-api/internal/services"
	"github.com/gin-gonic/gin"
)

type DownloadController struct {
	service *services.DownloadService
	repo    repositories.TaskRepository
}

func NewDownloadController(service *services.DownloadService, repo repositories.TaskRepository) *DownloadController {
	return &DownloadController{
		service: service,
		repo:    repo,
	}
}

// DownloadAsync godoc
// @Summary      Avvia download asincrono
// @Description  Avvia un task di download in background e restituisce immediatamente l'ID del task
// @Tags         Download
// @Accept       json
// @Produce      json
// @Param        request  body      models.DownloadRequest  true  "Parametri di download"
// @Success      202      {object}  map[string]interface{} "Contiene task_id, status e message"
// @Failure      400      {object}  map[string]string      "Richiesta malformata"
// @Router       /api/download [post]
func (ctrl *DownloadController) DownloadAsync(c *gin.Context) {
	var req models.DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setDefaults(&req)

	task, err := ctrl.service.DownloadAsync(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"task_id": task.ID,
		"status":  task.Status,
		"message": "Download task started successfully",
	})
}

// DownloadSync godoc
// @Summary      Avvia download sincrono
// @Description  Avvia il download e si blocca finché tutti i file non sono stati scaricati e taggati
// @Tags         Download
// @Accept       json
// @Produce      json
// @Param        request  body      models.DownloadRequest  true  "Parametri di download"
// @Success      200      {object}  map[string]interface{} "Contiene status e l'elenco dei file"
// @Failure      400      {object}  map[string]string      "Richiesta malformata"
// @Failure      500      {object}  map[string]string      "Errore durante il download"
// @Router       /api/download/sync [post]
func (ctrl *DownloadController) DownloadSync(c *gin.Context) {
	var req models.DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setDefaults(&req)

	files, err := ctrl.service.DownloadSync(req)
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
}

// GetStatus godoc
// @Summary      Traccia lo stato del task
// @Description  Restituisce lo stato attuale di un download asincrono specificando l'ID del task
// @Tags         Download
// @Produce      json
// @Param        id   path      string  true  "ID del task"
// @Success      200  {object}  models.Task
// @Failure      404  {object}  map[string]string "Task non trovato"
// @Router       /api/status/{id} [get]
func (ctrl *DownloadController) GetStatus(c *gin.Context) {
	id := c.Param("id")
	task, exists, err := ctrl.repo.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func setDefaults(req *models.DownloadRequest) {
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
}
