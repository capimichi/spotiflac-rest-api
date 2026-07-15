package main

import (
	"fmt"
	"os"

	"github.com/capimichi/spotiflac-rest-api/internal/controllers"
	"github.com/capimichi/spotiflac-rest-api/internal/repositories"
	"github.com/capimichi/spotiflac-rest-api/internal/services"
	"github.com/gin-gonic/gin"
)

// @title          SpotiFLAC REST API
// @version        1.0
// @description    A lightweight REST API wrapper for the SpotiFLAC backend.
// @host           localhost:8080
// @BasePath       /
func main() {
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	repo := repositories.NewInMemoryTaskRepository()
	svc := services.NewDownloadService(repo)

	healthCtrl := controllers.NewHealthController()
	downloadCtrl := controllers.NewDownloadController(svc, repo)

	r := controllers.SetupRouter(healthCtrl, downloadCtrl)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("SpotiFLAC REST API Server listening on port %s...\n", port)
	if err := r.Run(":" + port); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
	}
}
