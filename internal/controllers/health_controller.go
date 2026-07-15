package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// HealthCheck godoc
// @Summary      Health Check
// @Description  Verifica se il server API è attivo e funzionante
// @Tags         System
// @Produce      json
// @Success      200  {object}  map[string]interface{} "status ok"
// @Router       /api/health [get]
func (ctrl *HealthController) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "ok",
		"time":        time.Now().Format(time.RFC3339),
		"environment": gin.Mode(),
	})
}
