package main

import (
	"go-admin/internal/db"
	"go-admin/internal/routes"
	"log"
)

func main() {
	db.InitDB()
	defer db.CloseDB()

	r := routes.SetupRoutes()

	log.Println("Server starting on :8080")
	r.Run(":8080")
}
