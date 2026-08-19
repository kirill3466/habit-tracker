package repositories

import (
	"errors"

	"habit-tracker-api/internal/domain/models"
)

var ErrNotFound = errors.New("record not found")

type HabitRepository interface {
	Create(habit *models.Habit) error
	GetByID(id int64) (*models.Habit, error)
	GetAll(activeOnly bool) ([]*models.Habit, error)
	Update(habit *models.Habit) error
	Delete(id int64) error
}

type HabitLogRepository interface {
	Create(log *models.HabitLog) (*models.HabitLog, error)
	GetByHabitID(habitID int64) ([]*models.HabitLog, error)
	Delete(id int64) error
	GetStats(habitID int64) (*models.HabitStats, error)
}
