package postgres

import (
	"testing"
	"time"

	"habit-tracker-api/internal/domain/models"
)

func TestCalculateStats(t *testing.T) {
	now := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
	day := func(offset int) time.Time {
		return now.AddDate(0, 0, -offset)
	}

	tests := []struct {
		name  string
		dates []time.Time
		want  models.HabitStats
	}{
		{
			name:  "no logs",
			dates: nil,
			want:  models.HabitStats{HabitID: 1, TotalLogs: 0, CurrentStreak: 0, LongestStreak: 0, CompletionRate: 0},
		},
		{
			name:  "three consecutive days ending today",
			dates: []time.Time{day(2), day(1), day(0)},
			want:  models.HabitStats{HabitID: 1, TotalLogs: 3, CurrentStreak: 3, LongestStreak: 3, CompletionRate: 100},
		},
		{
			name:  "two consecutive days ending yesterday",
			dates: []time.Time{day(2), day(1)},
			want:  models.HabitStats{HabitID: 1, TotalLogs: 2, CurrentStreak: 2, LongestStreak: 2, CompletionRate: 100},
		},
		{
			name:  "last log too old, current streak reset",
			dates: []time.Time{day(3), day(2)},
			want:  models.HabitStats{HabitID: 1, TotalLogs: 2, CurrentStreak: 0, LongestStreak: 2, CompletionRate: 100},
		},
		{
			name:  "gap in the middle",
			dates: []time.Time{day(3), day(2), day(0)},
			want:  models.HabitStats{HabitID: 1, TotalLogs: 3, CurrentStreak: 1, LongestStreak: 2, CompletionRate: 75},
		},
		{
			name:  "unsorted input",
			dates: []time.Time{day(1), day(0), day(2)},
			want:  models.HabitStats{HabitID: 1, TotalLogs: 3, CurrentStreak: 3, LongestStreak: 3, CompletionRate: 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateStats(tt.want.HabitID, tt.dates, now)

			if got.TotalLogs != tt.want.TotalLogs {
				t.Errorf("TotalLogs = %d, want %d", got.TotalLogs, tt.want.TotalLogs)
			}
			if got.CurrentStreak != tt.want.CurrentStreak {
				t.Errorf("CurrentStreak = %d, want %d", got.CurrentStreak, tt.want.CurrentStreak)
			}
			if got.LongestStreak != tt.want.LongestStreak {
				t.Errorf("LongestStreak = %d, want %d", got.LongestStreak, tt.want.LongestStreak)
			}
			if got.CompletionRate != tt.want.CompletionRate {
				t.Errorf("CompletionRate = %v, want %v", got.CompletionRate, tt.want.CompletionRate)
			}
			if tt.want.TotalLogs > 0 {
				min, max := tt.dates[0], tt.dates[0]
				for _, d := range tt.dates {
					if d.Before(min) {
						min = d
					}
					if d.After(max) {
						max = d
					}
				}
				if !got.FirstLoggedDate.Equal(min) {
					t.Errorf("FirstLoggedDate = %v, want %v", got.FirstLoggedDate, min)
				}
				if !got.LastLoggedDate.Equal(max) {
					t.Errorf("LastLoggedDate = %v, want %v", got.LastLoggedDate, max)
				}
			}
		})
	}
}
