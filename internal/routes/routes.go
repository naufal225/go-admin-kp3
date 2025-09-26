package routes

import (
    "go-admin/internal/handlers"

    "github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
    r := gin.Default()

    dashboardHandler := handlers.NewDashboardHandler()

    api := r.Group("/api/v1")
    {
        api.GET("/admin/dashboard", dashboardHandler.GetDashboardStats)
    }

    return r
}