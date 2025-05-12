package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/koldun11/yartime/server/internal/service"
)

// Handler структура для обработчиков
type Handler struct {
	service service.Servicer
}

// NewHandler создаёт новый Handler
func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetClientConfig обрабатывает запрос на получение конфигурации клиента
func (h *Handler) GetClientConfig(c *gin.Context) {
	config, err := h.service.GetClientConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, config)
}

// TODO: нормальные модели
// TODO: responder??

// SetAllowedHours обрабатывает запрос на установку диапазона времени
func (h *Handler) SetAllowedHours(c *gin.Context) {
	var req struct {
		Start string `json:"start"`
		End   string `json:"end"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.service.SetAllowedHours(req.Start, req.End); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "allowed hours updated"})
}

// SetCronLine обрабатывает запрос на установку строки cron
func (h *Handler) SetCronLine(c *gin.Context) {
	var req struct {
		CronLine string `json:"cron_line"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.service.SetCronLine(req.CronLine); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "cron line updated"})
}

// SetDailyLimit обрабатывает запрос на установку дневного лимита
func (h *Handler) SetDailyLimit(c *gin.Context) {
	var req struct {
		DailyLimitMinutes int `json:"daily_limit_minutes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.service.SetDailyLimit(req.DailyLimitMinutes); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "daily limit updated"})
}
