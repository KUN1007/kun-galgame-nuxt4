package repository

import (
	"time"

	adminModel "kun-galgame-api/internal/admin/model"

	"gorm.io/gorm"
)

type UpdateRepository struct {
	db *gorm.DB
}

func NewUpdateRepository(db *gorm.DB) *UpdateRepository {
	return &UpdateRepository{db: db}
}

// ── History ─────────────────────────────

// CountHistory returns total update log count.
func (r *UpdateRepository) CountHistory() int64 {
	var total int64
	r.db.Model(&adminModel.UpdateLog{}).Count(&total)
	return total
}

// FindHistoryPaginated returns paginated update logs ordered by created DESC.
func (r *UpdateRepository) FindHistoryPaginated(page, limit int) []adminModel.UpdateLog {
	var logs []adminModel.UpdateLog
	r.db.Order("created DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&logs)
	return logs
}

// CreateHistory inserts a new update log.
func (r *UpdateRepository) CreateHistory(log *adminModel.UpdateLog) error {
	return r.db.Create(log).Error
}

// UpdateHistory patches an update log row.
func (r *UpdateRepository) UpdateHistory(id int, fields map[string]any) error {
	return r.db.Model(&adminModel.UpdateLog{}).Where("id = ?", id).
		Updates(fields).Error
}

// DeleteHistory deletes an update log by ID.
func (r *UpdateRepository) DeleteHistory(id int) {
	r.db.Delete(&adminModel.UpdateLog{}, id)
}

// ── Todo ────────────────────────────────

// CountTodos returns total todo count, optionally filtered by status.
func (r *UpdateRepository) CountTodos(status *int) int64 {
	var total int64
	q := r.db.Model(&adminModel.Todo{})
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	q.Count(&total)
	return total
}

// FindTodosPaginated returns paginated todos ordered by created DESC,
// optionally filtered by status.
func (r *UpdateRepository) FindTodosPaginated(page, limit int, status *int) []adminModel.Todo {
	var todos []adminModel.Todo
	q := r.db.Order("created DESC")
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	q.
		Offset((page - 1) * limit).Limit(limit).
		Find(&todos)
	return todos
}

// CreateTodo inserts a new todo. If status is 2 (completed), sets CompletedTime.
func (r *UpdateRepository) CreateTodo(todo *adminModel.Todo) error {
	if todo.Status == 2 {
		now := time.Now()
		todo.CompletedTime = &now
	}
	return r.db.Create(todo).Error
}

// UpdateTodo patches a todo row. If fields["status"] == 2 the caller already
// included completed_time; otherwise we clear it explicitly.
func (r *UpdateRepository) UpdateTodo(id int, fields map[string]any) error {
	return r.db.Model(&adminModel.Todo{}).Where("id = ?", id).
		Updates(fields).Error
}

// DeleteTodo deletes a todo by ID.
func (r *UpdateRepository) DeleteTodo(id int) {
	r.db.Delete(&adminModel.Todo{}, id)
}
