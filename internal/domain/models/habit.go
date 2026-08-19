package models

import "time"

type Habit struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	Icon        string    `json:"icon"`
	Frequency   string    `json:"frequency"`
	TargetDays  int       `json:"target_days"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type HabitLog struct {
	ID        int64     `json:"id"`
	HabitID   int64     `json:"habit_id"`
	Date      time.Time `json:"date"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
}

type HabitStats struct {
	HabitID         int64     `json:"habit_id"`
	TotalLogs       int       `json:"total_logs"`
	CurrentStreak   int       `json:"current_streak"`
	LongestStreak   int       `json:"longest_streak"`
	CompletionRate  float64   `json:"completion_rate"`
	LastLoggedDate  time.Time `json:"last_logged_date"`
	FirstLoggedDate time.Time `json:"first_logged_date"`
}
