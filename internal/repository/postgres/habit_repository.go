package postgres

import (
	"database/sql"

	"habit-tracker-api/internal/domain/models"
	"habit-tracker-api/internal/domain/repositories"
)

type HabitRepository struct {
	db *sql.DB
}

func NewHabitRepository(db *sql.DB) *HabitRepository {
	return &HabitRepository{db: db}
}

func (r *HabitRepository) Create(habit *models.Habit) error {
	query := `
		INSERT INTO habits (title, description, color, icon, frequency, target_days, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(
		query,
		habit.Title,
		habit.Description,
		habit.Color,
		habit.Icon,
		habit.Frequency,
		habit.TargetDays,
		habit.IsActive,
	).Scan(&habit.ID, &habit.CreatedAt, &habit.UpdatedAt)
}

func (r *HabitRepository) GetByID(id int64) (*models.Habit, error) {
	query := `
		SELECT id, title, description, color, icon, frequency, target_days, is_active, created_at, updated_at
		FROM habits
		WHERE id = $1`

	habit := &models.Habit{}
	err := r.db.QueryRow(query, id).Scan(
		&habit.ID,
		&habit.Title,
		&habit.Description,
		&habit.Color,
		&habit.Icon,
		&habit.Frequency,
		&habit.TargetDays,
		&habit.IsActive,
		&habit.CreatedAt,
		&habit.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, repositories.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return habit, nil
}

func (r *HabitRepository) GetAll(activeOnly bool) ([]*models.Habit, error) {
	query := `
		SELECT id, title, description, color, icon, frequency, target_days, is_active, created_at, updated_at
		FROM habits`
	if activeOnly {
		query += ` WHERE is_active = TRUE`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	habits := make([]*models.Habit, 0)
	for rows.Next() {
		habit := &models.Habit{}
		if err := rows.Scan(
			&habit.ID,
			&habit.Title,
			&habit.Description,
			&habit.Color,
			&habit.Icon,
			&habit.Frequency,
			&habit.TargetDays,
			&habit.IsActive,
			&habit.CreatedAt,
			&habit.UpdatedAt,
		); err != nil {
			return nil, err
		}
		habits = append(habits, habit)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return habits, nil
}

func (r *HabitRepository) Update(habit *models.Habit) error {
	query := `
		UPDATE habits
		SET title = $1, description = $2, color = $3, icon = $4,
		    frequency = $5, target_days = $6, is_active = $7
		WHERE id = $8
		RETURNING updated_at`

	err := r.db.QueryRow(
		query,
		habit.Title,
		habit.Description,
		habit.Color,
		habit.Icon,
		habit.Frequency,
		habit.TargetDays,
		habit.IsActive,
		habit.ID,
	).Scan(&habit.UpdatedAt)
	if err == sql.ErrNoRows {
		return repositories.ErrNotFound
	}
	return err
}

func (r *HabitRepository) Delete(id int64) error {
	result, err := r.db.Exec(`DELETE FROM habits WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return repositories.ErrNotFound
	}
	return nil
}
