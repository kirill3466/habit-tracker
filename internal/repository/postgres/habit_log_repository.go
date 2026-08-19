package postgres

import (
	"database/sql"
	"sort"
	"time"

	"habit-tracker-api/internal/domain/models"
	"habit-tracker-api/internal/domain/repositories"
)

type HabitLogRepository struct {
	db *sql.DB
}

func NewHabitLogRepository(db *sql.DB) *HabitLogRepository {
	return &HabitLogRepository{db: db}
}

func (r *HabitLogRepository) Create(log *models.HabitLog) (*models.HabitLog, error) {
	query := `
		INSERT INTO habit_logs (habit_id, date, notes)
		VALUES ($1, $2, $3)
		ON CONFLICT (habit_id, date) DO UPDATE SET notes = EXCLUDED.notes
		RETURNING id, created_at`

	err := r.db.QueryRow(query, log.HabitID, log.Date, log.Notes).Scan(&log.ID, &log.CreatedAt)
	if err != nil {
		return nil, err
	}
	return log, nil
}

func (r *HabitLogRepository) GetByHabitID(habitID int64) ([]*models.HabitLog, error) {
	query := `
		SELECT id, habit_id, date, notes, created_at
		FROM habit_logs
		WHERE habit_id = $1
		ORDER BY date DESC`

	rows, err := r.db.Query(query, habitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]*models.HabitLog, 0)
	for rows.Next() {
		log := &models.HabitLog{}
		if err := rows.Scan(&log.ID, &log.HabitID, &log.Date, &log.Notes, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *HabitLogRepository) Delete(id int64) error {
	result, err := r.db.Exec(`DELETE FROM habit_logs WHERE id = $1`, id)
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

func (r *HabitLogRepository) GetStats(habitID int64) (*models.HabitStats, error) {
	var exists bool
	if err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM habits WHERE id = $1)`, habitID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, repositories.ErrNotFound
	}

	rows, err := r.db.Query(`SELECT date FROM habit_logs WHERE habit_id = $1 ORDER BY date ASC`, habitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []time.Time
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return nil, err
		}
		dates = append(dates, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return calculateStats(habitID, dates, time.Now()), nil
}

func calculateStats(habitID int64, dates []time.Time, now time.Time) *models.HabitStats {
	stats := &models.HabitStats{HabitID: habitID}
	if len(dates) == 0 {
		return stats
	}

	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })

	stats.TotalLogs = len(dates)
	stats.FirstLoggedDate = dates[0]
	stats.LastLoggedDate = dates[len(dates)-1]

	longest, current := 1, 1
	for i := 1; i < len(dates); i++ {
		if dates[i].Sub(dates[i-1]) == 24*time.Hour {
			current++
		} else {
			current = 1
		}
		if current > longest {
			longest = current
		}
	}
	stats.LongestStreak = longest

	today := now.Truncate(24 * time.Hour)
	last := stats.LastLoggedDate.Truncate(24 * time.Hour)
	daysSinceLast := int(today.Sub(last).Hours() / 24)
	if daysSinceLast <= 1 {
		streak := 1
		for i := len(dates) - 1; i > 0; i-- {
			if dates[i].Sub(dates[i-1]) == 24*time.Hour {
				streak++
			} else {
				break
			}
		}
		stats.CurrentStreak = streak
	} else {
		stats.CurrentStreak = 0
	}

	totalDays := int(stats.LastLoggedDate.Sub(stats.FirstLoggedDate).Hours()/24) + 1
	if totalDays > 0 {
		stats.CompletionRate = float64(stats.TotalLogs) / float64(totalDays) * 100
	}

	return stats
}
