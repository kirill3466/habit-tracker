package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"habit-tracker-api/internal/domain/models"
	"habit-tracker-api/internal/domain/repositories"
)

type HabitLogHandler struct {
	logRepo   repositories.HabitLogRepository
	habitRepo repositories.HabitRepository
}

func NewHabitLogHandler(
	logRepo repositories.HabitLogRepository,
	habitRepo repositories.HabitRepository,
) *HabitLogHandler {
	return &HabitLogHandler{logRepo: logRepo, habitRepo: habitRepo}
}

type createLogRequest struct {
	Date  string `json:"date"`
	Notes string `json:"notes"`
}

func (h *HabitLogHandler) Create(w http.ResponseWriter, r *http.Request) {
	habitID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid habit id")
		return
	}

	if _, err := h.habitRepo.GetByID(habitID); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			respondError(w, http.StatusNotFound, "habit not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to fetch habit")
		return
	}

	var req createLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	logDate := time.Now()
	if req.Date != "" {
		parsed, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			respondError(w, http.StatusBadRequest, "date must be in YYYY-MM-DD format")
			return
		}
		logDate = parsed
	}

	log := &models.HabitLog{
		HabitID: habitID,
		Date:    logDate,
		Notes:   req.Notes,
	}

	created, err := h.logRepo.Create(log)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create log")
		return
	}
	respondJSON(w, http.StatusCreated, created)
}

func (h *HabitLogHandler) List(w http.ResponseWriter, r *http.Request) {
	habitID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid habit id")
		return
	}

	logs, err := h.logRepo.GetByHabitID(habitID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch logs")
		return
	}
	respondJSON(w, http.StatusOK, logs)
}

func (h *HabitLogHandler) Delete(w http.ResponseWriter, r *http.Request) {
	logID, err := strconv.ParseInt(r.PathValue("logID"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid log id")
		return
	}

	err = h.logRepo.Delete(logID)
	if errors.Is(err, repositories.ErrNotFound) {
		respondError(w, http.StatusNotFound, "log not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete log")
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}

func (h *HabitLogHandler) Stats(w http.ResponseWriter, r *http.Request) {
	habitID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid habit id")
		return
	}

	stats, err := h.logRepo.GetStats(habitID)
	if errors.Is(err, repositories.ErrNotFound) {
		respondError(w, http.StatusNotFound, "habit not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch stats")
		return
	}
	respondJSON(w, http.StatusOK, stats)
}
