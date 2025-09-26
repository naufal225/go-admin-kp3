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
    stats := h.service.GetDashboardStats()
    c.JSON(http.StatusOK, gin.H{
        "data": stats,
    })
}