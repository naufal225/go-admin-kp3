package routes

import (
	"go-admin/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// Setup CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true // Untuk development
	// config.AllowOrigins = []string{"http://localhost:8000"} // Untuk production
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	dashboardHandler := handlers.NewDashboardHandler()

	api := r.Group("/api/v1")
	{
		api.GET("/admin/dashboard", dashboardHandler.GetDashboardStats)
	}

	return r
}
