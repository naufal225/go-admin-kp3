package handlers

import (
	"go-admin/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service *services.DashboardService
}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{
		service: &services.DashboardService{},
	}
}

func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	period := c.Query("range")
	if period == "" {
		period = c.Query("periode")
	}
	stats := h.service.GetDashboardStats(period)
	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}
