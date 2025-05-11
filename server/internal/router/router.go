package router

import (
	"github.com/gin-gonic/gin"
	"github.com/koldun11/yartime/server/internal/handler"
)

// NewRouter создаёт и настраивает новый роутер Gin
func NewRouter(h *handler.Handler) *gin.Engine {
	r := gin.Default()

	// Маршруты
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
		})
	})

	r.GET("/client/config", h.GetClientConfig)
	r.POST("/client/allowed_hours", h.SetAllowedHours)
	r.POST("/client/cron_line", h.SetCronLine)
	r.POST("/client/daily_limit", h.SetDailyLimit)

	return r
}
