package api

import (
	"log"
	"net/http"
	"time"

	"habit-tracker-api/internal/domain/repositories"
)

func NewRouter(
	habitRepo repositories.HabitRepository,
	logRepo repositories.HabitLogRepository,
) http.Handler {
	habitHandler := NewHabitHandler(habitRepo)
	logHandler := NewHabitLogHandler(logRepo, habitRepo)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/habits", habitHandler.Create)
	mux.HandleFunc("GET /api/habits", habitHandler.List)
	mux.HandleFunc("GET /api/habits/{id}", habitHandler.GetByID)
	mux.HandleFunc("PATCH /api/habits/{id}", habitHandler.Update)
	mux.HandleFunc("DELETE /api/habits/{id}", habitHandler.Delete)

	mux.HandleFunc("POST /api/habits/{id}/logs", logHandler.Create)
	mux.HandleFunc("GET /api/habits/{id}/logs", logHandler.List)
	mux.HandleFunc("GET /api/habits/{id}/stats", logHandler.Stats)
	mux.HandleFunc("DELETE /api/logs/{logID}", logHandler.Delete)

	return logRequests(mux)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}
