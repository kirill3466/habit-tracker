package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"habit-tracker-api/internal/domain/models"
	"habit-tracker-api/internal/domain/repositories"
)

type HabitHandler struct {
	habitRepo repositories.HabitRepository
}

func NewHabitHandler(habitRepo repositories.HabitRepository) *HabitHandler {
	return &HabitHandler{habitRepo: habitRepo}
}

type createHabitRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	Frequency   string `json:"frequency"`
	TargetDays  *int   `json:"target_days"`
}

type updateHabitRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
	Icon        *string `json:"icon"`
	Frequency   *string `json:"frequency"`
	TargetDays  *int    `json:"target_days"`
	IsActive    *bool   `json:"is_active"`
}

func (h *HabitHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "title is required")
		return
	}

	frequency := req.Frequency
	if frequency == "" {
		frequency = "daily"
	}
	targetDays := 7
	if req.TargetDays != nil {
		targetDays = *req.TargetDays
	}

	habit := &models.Habit{
		Title:       req.Title,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		Frequency:   frequency,
		TargetDays:  targetDays,
		IsActive:    true,
	}

	if err := h.habitRepo.Create(habit); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create habit: "+err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, habit)
}

func (h *HabitHandler) List(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"
	habits, err := h.habitRepo.GetAll(activeOnly)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch habits")
		return
	}
	respondJSON(w, http.StatusOK, habits)
}

func (h *HabitHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid habit id")
		return
	}

	habit, err := h.habitRepo.GetByID(id)
	if errors.Is(err, repositories.ErrNotFound) {
		respondError(w, http.StatusNotFound, "habit not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch habit")
		return
	}
	respondJSON(w, http.StatusOK, habit)
}

func (h *HabitHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid habit id")
		return
	}

	var req updateHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	habit, err := h.habitRepo.GetByID(id)
	if errors.Is(err, repositories.ErrNotFound) {
		respondError(w, http.StatusNotFound, "habit not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch habit")
		return
	}

	if req.Title != nil {
		habit.Title = *req.Title
	}
	if req.Description != nil {
		habit.Description = *req.Description
	}
	if req.Color != nil {
		habit.Color = *req.Color
	}
	if req.Icon != nil {
		habit.Icon = *req.Icon
	}
	if req.Frequency != nil {
		habit.Frequency = *req.Frequency
	}
	if req.TargetDays != nil {
		habit.TargetDays = *req.TargetDays
	}
	if req.IsActive != nil {
		habit.IsActive = *req.IsActive
	}

	if err := h.habitRepo.Update(habit); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update habit")
		return
	}
	respondJSON(w, http.StatusOK, habit)
}

func (h *HabitHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid habit id")
		return
	}

	err = h.habitRepo.Delete(id)
	if errors.Is(err, repositories.ErrNotFound) {
		respondError(w, http.StatusNotFound, "habit not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete habit")
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}
