package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"habit-tracker-api/internal/domain/models"
	"habit-tracker-api/internal/domain/repositories"
)

type fakeHabitRepo struct {
	habits map[int64]*models.Habit
	nextID int64
}

func newFakeHabitRepo() *fakeHabitRepo {
	return &fakeHabitRepo{habits: map[int64]*models.Habit{}, nextID: 1}
}

func (f *fakeHabitRepo) Create(h *models.Habit) error {
	h.ID = f.nextID
	f.nextID++
	now := time.Now()
	h.CreatedAt = now
	h.UpdatedAt = now
	f.habits[h.ID] = h
	return nil
}

func (f *fakeHabitRepo) GetByID(id int64) (*models.Habit, error) {
	if h, ok := f.habits[id]; ok {
		return h, nil
	}
	return nil, repositories.ErrNotFound
}

func (f *fakeHabitRepo) GetAll(activeOnly bool) ([]*models.Habit, error) {
	habits := make([]*models.Habit, 0)
	for _, h := range f.habits {
		if activeOnly && !h.IsActive {
			continue
		}
		habits = append(habits, h)
	}
	return habits, nil
}

func (f *fakeHabitRepo) Update(h *models.Habit) error {
	if _, ok := f.habits[h.ID]; !ok {
		return repositories.ErrNotFound
	}
	h.UpdatedAt = time.Now()
	f.habits[h.ID] = h
	return nil
}

func (f *fakeHabitRepo) Delete(id int64) error {
	if _, ok := f.habits[id]; !ok {
		return repositories.ErrNotFound
	}
	delete(f.habits, id)
	return nil
}

type fakeLogRepo struct {
	logs   map[int64]*models.HabitLog
	nextID int64
}

func newFakeLogRepo() *fakeLogRepo {
	return &fakeLogRepo{logs: map[int64]*models.HabitLog{}, nextID: 1}
}

func (f *fakeLogRepo) Create(l *models.HabitLog) (*models.HabitLog, error) {
	l.ID = f.nextID
	f.nextID++
	l.CreatedAt = time.Now()
	f.logs[l.ID] = l
	return l, nil
}

func (f *fakeLogRepo) GetByHabitID(habitID int64) ([]*models.HabitLog, error) {
	logs := make([]*models.HabitLog, 0)
	for _, l := range f.logs {
		if l.HabitID == habitID {
			logs = append(logs, l)
		}
	}
	return logs, nil
}

func (f *fakeLogRepo) Delete(id int64) error {
	if _, ok := f.logs[id]; !ok {
		return repositories.ErrNotFound
	}
	delete(f.logs, id)
	return nil
}

func (f *fakeLogRepo) GetStats(habitID int64) (*models.HabitStats, error) {
	return &models.HabitStats{HabitID: habitID}, nil
}

func newTestRouter() (http.Handler, *fakeHabitRepo, *fakeLogRepo) {
	habitRepo := newFakeHabitRepo()
	logRepo := newFakeLogRepo()
	return NewRouter(habitRepo, logRepo), habitRepo, logRepo
}

func doRequest(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}
func TestHealth(t *testing.T) {
	router, _, _ := newTestRouter()
	rec := doRequest(router, http.MethodGet, "/health", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCreateHabit(t *testing.T) {
	router, _, _ := newTestRouter()
	rec := doRequest(router, http.MethodPost, "/api/habits", `{"title": "Читать", "frequency": "daily", "target_days": 30}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var habit models.Habit
	if err := json.Unmarshal(rec.Body.Bytes(), &habit); err != nil {
		t.Fatal(err)
	}
	if habit.ID == 0 {
		t.Error("expected generated id")
	}
	if habit.Title != "Читать" {
		t.Errorf("title = %q, want Читать", habit.Title)
	}
	if habit.Frequency != "daily" {
		t.Errorf("frequency = %q, want daily", habit.Frequency)
	}
	if !habit.IsActive {
		t.Error("expected is_active=true by default")
	}
}

func TestCreateHabitMissingTitle(t *testing.T) {
	router, _, _ := newTestRouter()
	rec := doRequest(router, http.MethodPost, "/api/habits", `{"description": "no title"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetHabitNotFound(t *testing.T) {
	router, _, _ := newTestRouter()
	rec := doRequest(router, http.MethodGet, "/api/habits/999", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUpdateHabit(t *testing.T) {
	router, habitRepo, _ := newTestRouter()
	habit := &models.Habit{Title: "Old"}
	if err := habitRepo.Create(habit); err != nil {
		t.Fatal(err)
	}

	rec := doRequest(router, http.MethodPatch, fmt.Sprintf("/api/habits/%d", habit.ID), `{"title": "New", "is_active": false}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var updated models.Habit
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Title != "New" {
		t.Errorf("title = %q, want New", updated.Title)
	}
	if updated.IsActive {
		t.Error("expected is_active=false")
	}
}

func TestDeleteHabit(t *testing.T) {
	router, habitRepo, _ := newTestRouter()
	habit := &models.Habit{Title: "Temp"}
	if err := habitRepo.Create(habit); err != nil {
		t.Fatal(err)
	}

	rec := doRequest(router, http.MethodDelete, fmt.Sprintf("/api/habits/%d", habit.ID), "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	if _, err := habitRepo.GetByID(habit.ID); err != repositories.ErrNotFound {
		t.Errorf("expected habit to be deleted, got err=%v", err)
	}
}

func TestCreateLogInvalidDate(t *testing.T) {
	router, habitRepo, _ := newTestRouter()
	habit := &models.Habit{Title: "Test"}
	if err := habitRepo.Create(habit); err != nil {
		t.Fatal(err)
	}

	rec := doRequest(router, http.MethodPost, fmt.Sprintf("/api/habits/%d/logs", habit.ID), `{"date": "19-08-2026"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateLogForMissingHabit(t *testing.T) {
	router, _, _ := newTestRouter()
	rec := doRequest(router, http.MethodPost, "/api/habits/777/logs", `{}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestListHabits(t *testing.T) {
	router, habitRepo, _ := newTestRouter()
	for _, title := range []string{"A", "B"} {
		if err := habitRepo.Create(&models.Habit{Title: title}); err != nil {
			t.Fatal(err)
		}
	}

	rec := doRequest(router, http.MethodGet, "/api/habits", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var habits []models.Habit
	if err := json.Unmarshal(rec.Body.Bytes(), &habits); err != nil {
		t.Fatal(err)
	}
	if len(habits) != 2 {
		t.Errorf("expected 2 habits, got %d", len(habits))
	}
}

