package main

import (
	"log"
	"net/http"

	"habit-tracker-api/internal/api"
	"habit-tracker-api/internal/config"
	"habit-tracker-api/internal/database"
	"habit-tracker-api/internal/repository/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.Connect(cfg.Database.DSN()); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	habitRepo := postgres.NewHabitRepository(database.DB)
	logRepo := postgres.NewHabitLogRepository(database.DB)

	router := api.NewRouter(habitRepo, logRepo)

	log.Printf("🚀 API server starting on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
