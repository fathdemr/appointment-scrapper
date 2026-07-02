package handler

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var startTime = time.Now()

// SystemHandler sistem geneli endpoint'leri yönetir.
type SystemHandler struct {
	logger *zap.Logger
}

func NewSystemHandler(logger *zap.Logger) *SystemHandler {
	return &SystemHandler{logger: logger}
}

// Health godoc
// @Summary      Sağlık kontrolü
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func (h *SystemHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// SystemStatus godoc
// @Summary      Sistem durumu
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/system/status [get]
func (h *SystemHandler) SystemStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":     "ok",
		"uptime":     time.Since(startTime).String(),
		"goroutines": runtime.NumGoroutine(),
	})
}
